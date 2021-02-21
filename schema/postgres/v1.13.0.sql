ALTER TABLE "visitor_stats" ADD COLUMN "views" integer NOT NULL DEFAULT 0;
ALTER TABLE "visitor_stats" ADD COLUMN "average_session_duration_seconds" integer NOT NULL DEFAULT 0;
