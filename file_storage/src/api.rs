use axum::extract::{Path, Query, State};
use axum::http::header::CONTENT_TYPE;
use axum::http::{HeaderMap, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::Json;
use bytes::Bytes;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::error::AppError;
use crate::storage::Storage;

#[derive(Clone)]
pub struct AppState {
    pub storage: Storage,
}

#[derive(Serialize)]
pub(crate) struct UploadResponse {
    uuid: String,
}

#[derive(Serialize)]
pub(crate) struct HashResponse {
    sha256: String,
}

#[derive(Deserialize)]
pub struct PatchParams {
    pub offset: u64,
}

#[derive(Deserialize)]
pub struct HashParams {
    pub offset: Option<u64>,
    pub length: Option<u64>,
}

pub async fn upload(
    State(state): State<AppState>,
    headers: HeaderMap,
    body: Bytes,
) -> Result<(StatusCode, Json<UploadResponse>), AppError> {
    let data = match headers
        .get(CONTENT_TYPE)
        .and_then(|v| v.to_str().ok())
    {
        Some(ct) if ct.starts_with("multipart/form-data") => {
            extract_multipart_field(&body, ct, "file")?
        }
        _ => body.to_vec(),
    };

    let uuid = Uuid::now_v7();
    state.storage.store(&uuid, &data).await?;
    Ok((
        StatusCode::CREATED,
        Json(UploadResponse {
            uuid: uuid.to_string(),
        }),
    ))
}

pub async fn get_file(
    State(state): State<AppState>,
    Path(uuid_str): Path<String>,
) -> Result<Response, AppError> {
    let uuid = Uuid::parse_str(&uuid_str).map_err(|_| AppError::InvalidUuid)?;
    let data = state.storage.get(&uuid).await?;
    Ok((
        [("content-type", "application/octet-stream")],
        Bytes::from(data),
    )
        .into_response())
}

pub async fn patch_file(
    State(state): State<AppState>,
    Path(uuid_str): Path<String>,
    Query(params): Query<PatchParams>,
    body: Bytes,
) -> Result<StatusCode, AppError> {
    let uuid = Uuid::parse_str(&uuid_str).map_err(|_| AppError::InvalidUuid)?;
    state.storage.patch(&uuid, params.offset, &body).await?;
    Ok(StatusCode::NO_CONTENT)
}

pub async fn hash_file(
    State(state): State<AppState>,
    Path(uuid_str): Path<String>,
    Query(params): Query<HashParams>,
) -> Result<Json<HashResponse>, AppError> {
    let uuid = Uuid::parse_str(&uuid_str).map_err(|_| AppError::InvalidUuid)?;
    let offset = params.offset.unwrap_or(0);
    let length = params.length.unwrap_or(u64::MAX);
    let hash = state.storage.hash(&uuid, offset, length).await?;
    Ok(Json(HashResponse { sha256: hash }))
}

pub async fn delete_file(
    State(state): State<AppState>,
    Path(uuid_str): Path<String>,
) -> Result<StatusCode, AppError> {
    let uuid = Uuid::parse_str(&uuid_str).map_err(|_| AppError::InvalidUuid)?;
    state.storage.delete(&uuid).await?;
    Ok(StatusCode::NO_CONTENT)
}

fn extract_multipart_field(body: &[u8], content_type: &str, field_name: &str) -> Result<Vec<u8>, AppError> {
    let boundary = content_type
        .split("boundary=")
        .nth(1)
        .ok_or(AppError::InvalidMultipart)?
        .split(';')
        .next()
        .ok_or(AppError::InvalidMultipart)?
        .trim()
        .trim_matches('"');

    let boundary_start = format!("--{}", boundary);
    let boundary_bytes = boundary_start.as_bytes();
    let mut remaining = body;

    loop {
        let idx = remaining
            .windows(boundary_bytes.len())
            .position(|w| w == boundary_bytes);

        match idx {
            None => break,
            Some(pos) => {
                remaining = &remaining[pos + boundary_bytes.len()..];

                if remaining.starts_with(b"--") {
                    break;
                }
                if remaining.starts_with(b"\r\n") {
                    remaining = &remaining[2..];
                }

                let hdr_end = remaining.windows(4).position(|w| w == b"\r\n\r\n");

                match hdr_end {
                    None => continue,
                    Some(end) => {
                        let hdrs = std::str::from_utf8(&remaining[..end])
                            .map_err(|_| AppError::InvalidMultipart)?;

                        remaining = &remaining[end + 4..];

                        if hdrs.contains(&format!("name=\"{}\"", field_name)) {
                            let next = remaining
                                .windows(boundary_bytes.len())
                                .position(|w| w == boundary_bytes)
                                .unwrap_or(remaining.len());

                            let part_body = &remaining[..next];
                            let trimmed = if part_body.ends_with(b"\r\n") {
                                &part_body[..part_body.len() - 2]
                            } else {
                                part_body
                            };
                            return Ok(trimmed.to_vec());
                        }
                    }
                }
            }
        }
    }

    Err(AppError::MissingField)
}
