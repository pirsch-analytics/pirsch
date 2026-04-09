ALTER TABLE "session" {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} ADD COLUMN "channel" LowCardinality(String);
ALTER TABLE "page_view" {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} ADD COLUMN "channel" LowCardinality(String);
ALTER TABLE "event" {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} ADD COLUMN "channel" LowCardinality(String);
