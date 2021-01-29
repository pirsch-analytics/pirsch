ALTER TABLE "hit" ADD COLUMN "referrer_name" varchar(200);
ALTER TABLE "hit" ADD COLUMN "referrer_icon" varchar(2000);
ALTER TABLE "referrer_stats" ADD COLUMN "referrer_name" varchar(200);
ALTER TABLE "referrer_stats" ADD COLUMN "referrer_icon" varchar(2000);
ALTER TABLE "referrer_stats" ADD COLUMN "bounces" integer NOT NULL DEFAULT 0;
