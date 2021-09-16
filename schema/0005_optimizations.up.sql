ALTER TABLE "hit" RENAME COLUMN "previous_time_on_page_seconds" TO "duration_seconds";
ALTER TABLE "hit" ADD COLUMN "is_bounce" UInt8 DEFAULT 1;
ALTER TABLE "hit" ADD COLUMN "entry_path" String DEFAULT '';
ALTER TABLE "hit" ADD COLUMN "page_views" UInt16 DEFAULT 1;

ALTER TABLE "event" RENAME COLUMN "previous_time_on_page_seconds" TO "duration_seconds";
ALTER TABLE "event" ADD COLUMN "is_bounce" UInt8 DEFAULT 1;
ALTER TABLE "event" ADD COLUMN "entry_path" String DEFAULT '';
ALTER TABLE "event" ADD COLUMN "page_views" UInt16 DEFAULT 1;
