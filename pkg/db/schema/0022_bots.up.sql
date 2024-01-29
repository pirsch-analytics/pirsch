DROP TABLE "user_agent";
RENAME TABLE "bot" TO "request";
ALTER TABLE "request" ADD COLUMN "bot" Boolean DEFAULT 1;
