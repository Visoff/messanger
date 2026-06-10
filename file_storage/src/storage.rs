use std::ops::Range;
use std::path::PathBuf;
use std::sync::Arc;

use dashmap::DashMap;
use parking_lot::Mutex;
use sha2::{Digest, Sha256};
use tokio::fs;
use tokio::io::{AsyncReadExt, AsyncSeekExt, AsyncWriteExt};
use uuid::Uuid;

use crate::error::AppError;

#[derive(Clone)]
pub struct Storage {
    path: PathBuf,
    locks: Arc<DashMap<Uuid, Mutex<Vec<Range<u64>>>>>,
}

struct RangeGuard<'a> {
    locks: &'a Arc<DashMap<Uuid, Mutex<Vec<Range<u64>>>>>,
    uuid: Uuid,
    range: Range<u64>,
}

impl Drop for RangeGuard<'_> {
    fn drop(&mut self) {
        if let Some(entry) = self.locks.get(&self.uuid) {
            entry.value().lock().retain(|r| r != &self.range);
        }
    }
}

impl Storage {
    pub fn new(path: impl Into<PathBuf>) -> Self {
        Self {
            path: path.into(),
            locks: Arc::new(DashMap::new()),
        }
    }

    pub async fn init(&self) -> std::io::Result<()> {
        fs::create_dir_all(&self.path).await
    }

    fn file_path(&self, uuid: &Uuid) -> PathBuf {
        self.path.join(uuid.to_string())
    }

    pub async fn store(&self, uuid: &Uuid, data: &[u8]) -> Result<(), AppError> {
        fs::write(self.file_path(uuid), data).await?;
        Ok(())
    }

    pub async fn get(&self, uuid: &Uuid) -> Result<Vec<u8>, AppError> {
        Ok(fs::read(self.file_path(uuid)).await?)
    }

    pub async fn delete(&self, uuid: &Uuid) -> Result<(), AppError> {
        let path = self.file_path(uuid);
        fs::remove_file(&path).await?;
        self.locks.remove(uuid);
        Ok(())
    }

    pub async fn patch(&self, uuid: &Uuid, offset: u64, data: &[u8]) -> Result<(), AppError> {
        let path = self.file_path(uuid);
        let len = data.len() as u64;
        let new_range = offset..offset + len;

        let guard = {
            let entry = self.locks.entry(*uuid).or_insert_with(|| Mutex::new(Vec::new()));
            let mut ranges = entry.value().lock();
            if ranges.iter().any(|r| ranges_intersect(r, &new_range)) {
                return Err(AppError::Conflict);
            }
            ranges.push(new_range.clone());
            RangeGuard {
                locks: &self.locks,
                uuid: *uuid,
                range: new_range,
            }
        };

        let mut file = fs::OpenOptions::new()
            .create(true)
            .write(true)
            .open(&path)
            .await?;
        file.seek(tokio::io::SeekFrom::Start(offset)).await?;
        file.write_all(data).await?;

        drop(guard);
        Ok(())
    }

    pub async fn hash(
        &self,
        uuid: &Uuid,
        offset: u64,
        length: u64,
    ) -> Result<String, AppError> {
        let path = self.file_path(uuid);
        let mut file = fs::File::open(&path).await?;
        let file_len = file.metadata().await?.len();

        let start = offset.min(file_len);
        let end = if length == u64::MAX {
            file_len
        } else {
            (start + length).min(file_len)
        };

        file.seek(tokio::io::SeekFrom::Start(start)).await?;

        let mut hasher = Sha256::new();
        let mut buf = [0u8; 8192];
        let mut remaining = end - start;
        while remaining > 0 {
            let to_read = buf.len().min(remaining as usize);
            let n = file.read(&mut buf[..to_read]).await?;
            if n == 0 {
                break;
            }
            hasher.update(&buf[..n]);
            remaining -= n as u64;
        }

        Ok(format!("{:x}", hasher.finalize()))
    }
}

fn ranges_intersect(a: &Range<u64>, b: &Range<u64>) -> bool {
    a.start < b.end && b.start < a.end
}
