use std::net::SocketAddr;

use axum::http::{Method, Request, StatusCode};
use axum::middleware::{self, Next};
use axum::response::Response;
use axum::routing::{get, post};
use axum::Router;
use tower_http::cors::CorsLayer;
use tracing_subscriber::EnvFilter;

mod api;
mod error;
mod storage;

use api::AppState;
use storage::Storage;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let store_path = std::env::var("STORAGE_PATH").unwrap_or_else(|_| "./store".into());
    let bind_addr = std::env::var("BIND_ADDR").unwrap_or_else(|_| "0.0.0.0:3001".into());

    let storage = Storage::new(&store_path);
    storage.init().await.expect("failed to init storage");

    let state = AppState { storage };

    async fn auth_middleware(
        req: Request<axum::body::Body>,
        next: Next,
    ) -> Result<Response, StatusCode> {
        // Allow GET/HEAD without API key (public avatar URLs)
        if req.method() == Method::GET || req.method() == Method::HEAD {
            return Ok(next.run(req).await);
        }

        let api_key = std::env::var("FILE_STORAGE_API_KEY").ok();
        match api_key {
            Some(key) => {
                let header = req
                    .headers()
                    .get("X-Api-Key")
                    .and_then(|v| v.to_str().ok());
                match header {
                    Some(h) if h == key => Ok(next.run(req).await),
                    _ => Err(StatusCode::UNAUTHORIZED),
                }
            }
            None => Ok(next.run(req).await),
        }
    }

    let app = Router::new()
        .route("/file", post(api::upload))
        .route(
            "/{uuid}",
            get(api::get_file)
                .patch(api::patch_file)
                .delete(api::delete_file),
        )
        .route("/{uuid}/hash", get(api::hash_file))
        .layer(middleware::from_fn(auth_middleware))
        .layer(CorsLayer::permissive())
        .with_state(state);

    let addr: SocketAddr = bind_addr.parse().expect("invalid BIND_ADDR");
    tracing::info!("listening on {addr}");

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
