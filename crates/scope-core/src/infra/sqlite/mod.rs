use async_trait::async_trait;
use sqlx::sqlite::{SqliteConnectOptions, SqlitePoolOptions};
use sqlx::{Row, SqlitePool};

use crate::domain::{
    Fort, Notification, NotificationLevel, ServiceConfig, Store, Urgency,
};
use crate::Error;

pub struct SqliteStore {
    pool: SqlitePool,
}

impl SqliteStore {
    pub async fn open(path: &str) -> Result<Self, Error> {
        let opts = if path == ":memory:" {
            SqliteConnectOptions::new()
                .filename(":memory:")
                .create_if_missing(true)
        } else {
            SqliteConnectOptions::new()
                .filename(path)
                .create_if_missing(true)
                .pragma("journal_mode", "WAL")
                .pragma("busy_timeout", "5000")
                .pragma("synchronous", "NORMAL")
        };

        let pool = SqlitePoolOptions::new()
            .max_connections(1)
            .connect_with(opts)
            .await?;

        // Enable foreign keys
        sqlx::query("PRAGMA foreign_keys = ON")
            .execute(&pool)
            .await?;

        // Run migration
        let migration = include_str!("../../../migrations/sqlite/001_initial.sql");
        for statement in migration.split(';') {
            let stmt = statement.trim();
            if !stmt.is_empty() {
                sqlx::query(stmt).execute(&pool).await?;
            }
        }

        Ok(Self { pool })
    }
}

fn urgency_to_str(u: Urgency) -> &'static str {
    match u {
        Urgency::Passive => "passive",
        Urgency::Active => "active",
    }
}

fn str_to_urgency(s: &str) -> Urgency {
    match s {
        "active" => Urgency::Active,
        _ => Urgency::Passive,
    }
}

fn level_to_str(l: NotificationLevel) -> &'static str {
    match l {
        NotificationLevel::Mute => "mute",
        NotificationLevel::PassiveOnly => "passive_only",
        NotificationLevel::AllowUrgent => "allow_urgent",
    }
}

fn str_to_level(s: &str) -> NotificationLevel {
    match s {
        "mute" => NotificationLevel::Mute,
        "passive_only" => NotificationLevel::PassiveOnly,
        _ => NotificationLevel::AllowUrgent,
    }
}

#[async_trait]
impl Store for SqliteStore {
    async fn list_forts(&self) -> crate::domain::ports::Result<Vec<Fort>> {
        let rows = sqlx::query("SELECT name, local, gateway, active FROM forts")
            .fetch_all(&self.pool)
            .await?;

        let mut forts = Vec::with_capacity(rows.len());
        for row in rows {
            let name: String = row.get("name");
            let svc_rows =
                sqlx::query("SELECT url FROM fort_services WHERE fort_name = ?")
                    .bind(&name)
                    .fetch_all(&self.pool)
                    .await?;
            let services = svc_rows
                .iter()
                .map(|r| ServiceConfig {
                    url: r.get("url"),
                })
                .collect();
            forts.push(Fort {
                name,
                local: row.get::<i32, _>("local") != 0,
                gateway: row.get("gateway"),
                services,
            });
        }
        Ok(forts)
    }

    async fn get_fort(&self, name: &str) -> crate::domain::ports::Result<Fort> {
        let row =
            sqlx::query("SELECT name, local, gateway, active FROM forts WHERE name = ?")
                .bind(name)
                .fetch_optional(&self.pool)
                .await?
                .ok_or_else(|| Error::NotFound(format!("fort '{name}'")))?;

        let svc_rows =
            sqlx::query("SELECT url FROM fort_services WHERE fort_name = ?")
                .bind(name)
                .fetch_all(&self.pool)
                .await?;
        let services = svc_rows
            .iter()
            .map(|r| ServiceConfig {
                url: r.get("url"),
            })
            .collect();

        Ok(Fort {
            name: row.get("name"),
            local: row.get::<i32, _>("local") != 0,
            gateway: row.get("gateway"),
            services,
        })
    }

    async fn upsert_fort(&self, fort: &Fort) -> crate::domain::ports::Result<()> {
        sqlx::query(
            "INSERT OR REPLACE INTO forts (name, local, gateway) VALUES (?, ?, ?)",
        )
        .bind(&fort.name)
        .bind(fort.local as i32)
        .bind(&fort.gateway)
        .execute(&self.pool)
        .await?;

        // Replace services: delete old, insert new
        sqlx::query("DELETE FROM fort_services WHERE fort_name = ?")
            .bind(&fort.name)
            .execute(&self.pool)
            .await?;

        for svc in &fort.services {
            sqlx::query(
                "INSERT INTO fort_services (fort_name, url) VALUES (?, ?)",
            )
            .bind(&fort.name)
            .bind(&svc.url)
            .execute(&self.pool)
            .await?;
        }

        Ok(())
    }

    async fn delete_fort(&self, name: &str) -> crate::domain::ports::Result<()> {
        let result = sqlx::query("DELETE FROM forts WHERE name = ?")
            .bind(name)
            .execute(&self.pool)
            .await?;

        if result.rows_affected() == 0 {
            return Err(Error::NotFound(format!("fort '{name}'")));
        }
        Ok(())
    }

    async fn get_active_fort(&self) -> crate::domain::ports::Result<Option<String>> {
        let row =
            sqlx::query("SELECT name FROM forts WHERE active = 1")
                .fetch_optional(&self.pool)
                .await?;
        Ok(row.map(|r| r.get("name")))
    }

    async fn set_active_fort(&self, name: &str) -> crate::domain::ports::Result<()> {
        // Verify fort exists
        let exists =
            sqlx::query("SELECT 1 FROM forts WHERE name = ?")
                .bind(name)
                .fetch_optional(&self.pool)
                .await?;
        if exists.is_none() {
            return Err(Error::NotFound(format!("fort '{name}'")));
        }

        sqlx::query("UPDATE forts SET active = 0")
            .execute(&self.pool)
            .await?;
        sqlx::query("UPDATE forts SET active = 1 WHERE name = ?")
            .bind(name)
            .execute(&self.pool)
            .await?;
        Ok(())
    }

    async fn insert_notification(
        &self,
        n: &Notification,
    ) -> crate::domain::ports::Result<i64> {
        let created = n.created_at.to_rfc3339();
        let row = sqlx::query(
            "INSERT INTO notifications (service, title, body, urgency, route, created_at) \
             VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
        )
        .bind(&n.service)
        .bind(&n.title)
        .bind(&n.body)
        .bind(urgency_to_str(n.urgency))
        .bind(&n.route)
        .bind(&created)
        .fetch_one(&self.pool)
        .await?;

        Ok(row.get("id"))
    }

    async fn list_notifications(
        &self,
        limit: i64,
        before_id: Option<i64>,
    ) -> crate::domain::ports::Result<Vec<Notification>> {
        let rows = if let Some(bid) = before_id {
            sqlx::query(
                "SELECT id, service, title, body, urgency, route, read, created_at \
                 FROM notifications WHERE id < ? ORDER BY id DESC LIMIT ?",
            )
            .bind(bid)
            .bind(limit)
            .fetch_all(&self.pool)
            .await?
        } else {
            sqlx::query(
                "SELECT id, service, title, body, urgency, route, read, created_at \
                 FROM notifications ORDER BY id DESC LIMIT ?",
            )
            .bind(limit)
            .fetch_all(&self.pool)
            .await?
        };

        let mut notifications = Vec::with_capacity(rows.len());
        for row in rows {
            let created_str: String = row.get("created_at");
            let created_at = chrono::DateTime::parse_from_rfc3339(&created_str)
                .map(|dt| dt.with_timezone(&chrono::Utc))
                .unwrap_or_else(|_| chrono::Utc::now());

            notifications.push(Notification {
                id: Some(row.get("id")),
                service: row.get("service"),
                title: row.get("title"),
                body: row.get("body"),
                urgency: str_to_urgency(row.get("urgency")),
                route: row.get("route"),
                read: row.get::<i32, _>("read") != 0,
                created_at,
            });
        }
        Ok(notifications)
    }

    async fn unread_count(&self) -> crate::domain::ports::Result<i64> {
        let row =
            sqlx::query("SELECT COUNT(*) as cnt FROM notifications WHERE read = 0")
                .fetch_one(&self.pool)
                .await?;
        Ok(row.get::<i64, _>("cnt"))
    }

    async fn mark_read(&self, up_to_id: i64) -> crate::domain::ports::Result<()> {
        sqlx::query("UPDATE notifications SET read = 1 WHERE id <= ? AND read = 0")
            .bind(up_to_id)
            .execute(&self.pool)
            .await?;
        Ok(())
    }

    async fn get_preference(
        &self,
        service: &str,
    ) -> crate::domain::ports::Result<NotificationLevel> {
        let row =
            sqlx::query("SELECT level FROM preferences WHERE service = ?")
                .bind(service)
                .fetch_optional(&self.pool)
                .await?;
        match row {
            Some(r) => Ok(str_to_level(r.get("level"))),
            None => Ok(NotificationLevel::AllowUrgent),
        }
    }

    async fn set_preference(
        &self,
        service: &str,
        level: NotificationLevel,
    ) -> crate::domain::ports::Result<()> {
        sqlx::query(
            "INSERT OR REPLACE INTO preferences (service, level) VALUES (?, ?)",
        )
        .bind(service)
        .bind(level_to_str(level))
        .execute(&self.pool)
        .await?;
        Ok(())
    }
}

#[cfg(test)]
mod tests;
