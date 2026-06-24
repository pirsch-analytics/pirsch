ALTER TABLE "session" ADD COLUMN hostname String;
ALTER TABLE "page_view" ADD COLUMN hostname String;
ALTER TABLE "event" ADD COLUMN hostname String;
ALTER TABLE "request" ADD COLUMN hostname String;
