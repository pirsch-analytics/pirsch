ALTER TABLE hit ADD COLUMN visitor_id UInt64;
ALTER TABLE event ADD COLUMN visitor_id UInt64;
ALTER TABLE hit UPDATE visitor_id = sipHash64(fingerprint) WHERE 1;
ALTER TABLE event UPDATE visitor_id = sipHash64(fingerprint) WHERE 1;
ALTER TABLE hit DROP COLUMN fingerprint;
ALTER TABLE event DROP COLUMN fingerprint;
