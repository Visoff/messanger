use std::net::SocketAddr;

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

    let app = Router::new()
        .route("/file", post(api::upload))
        .route(
            "/{uuid}",
            get(api::get_file)
                .patch(api::patch_file)
                .delete(api::delete_file),
        )
        .route("/{uuid}/hash", get(api::hash_file))
        .layer(CorsLayer::permissive())
        .with_state(state);

    let addr: SocketAddr = bind_addr.parse().expect("invalid BIND_ADDR");
    tracing::info!("listening on {addr}");

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
