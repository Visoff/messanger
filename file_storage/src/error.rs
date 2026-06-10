use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde_json::json;

use tracing;

#[derive(Debug)]
pub enum AppError {
    NotFound,
    Conflict,
    InvalidUuid,
    MissingField,
    InvalidMultipart,
    Storage(String),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound => (StatusCode::NOT_FOUND, "not found"),
            AppError::Conflict => (StatusCode::CONFLICT, "range conflict"),
            AppError::InvalidUuid => (StatusCode::BAD_REQUEST, "invalid uuid"),
            AppError::MissingField => {
                (StatusCode::BAD_REQUEST, "missing 'file' field in multipart")
            }
            AppError::InvalidMultipart => (StatusCode::BAD_REQUEST, "invalid multipart data"),
            AppError::Storage(msg) => {
                tracing::warn!("storage error: {msg}");
                (StatusCode::INTERNAL_SERVER_ERROR, "storage error")
            }
        };
        (status, Json(json!({"error": message}))).into_response()
    }
}

impl From<std::io::Error> for AppError {
    fn from(e: std::io::Error) -> Self {
        match e.kind() {
            std::io::ErrorKind::NotFound => AppError::NotFound,
            _ => AppError::Storage(e.to_string()),
        }
    }
}
