ALTER TABLE "visitor_time_stats" DROP COLUMN "path";
ALTER TABLE "visitor_stats" ALTER COLUMN "path" DROP NOT NULL;
ALTER TABLE "language_stats" ALTER COLUMN "path" DROP NOT NULL;
ALTER TABLE "referrer_stats" ALTER COLUMN "path" DROP NOT NULL;
ALTER TABLE "os_stats" ALTER COLUMN "path" DROP NOT NULL;
ALTER TABLE "browser_stats" ALTER COLUMN "path" DROP NOT NULL;
