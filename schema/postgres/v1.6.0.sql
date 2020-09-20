ALTER TABLE "hit" ADD COLUMN "country_code" character varying(2);
ALTER TABLE "hit" ADD COLUMN "screen_width" integer DEFAULT 0;
ALTER TABLE "hit" ADD COLUMN "screen_height" integer DEFAULT 0;

CREATE TABLE "screen_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    visitors integer NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL
);

CREATE SEQUENCE screen_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE screen_stats_id_seq OWNED BY "screen_stats".id;
ALTER TABLE ONLY "screen_stats" ALTER COLUMN id SET DEFAULT nextval('screen_stats_id_seq'::regclass);
ALTER TABLE ONLY "screen_stats" ADD CONSTRAINT screen_stats_pkey PRIMARY KEY (id);
CREATE INDEX screen_stats_day_index ON screen_stats(day);

CREATE TABLE "country_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    visitors integer NOT NULL,
    country_code character varying(2)
);

CREATE SEQUENCE country_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE country_stats_id_seq OWNED BY "country_stats".id;
ALTER TABLE ONLY "country_stats" ALTER COLUMN id SET DEFAULT nextval('country_stats_id_seq'::regclass);
ALTER TABLE ONLY "country_stats" ADD CONSTRAINT country_stats_pkey PRIMARY KEY (id);
CREATE INDEX country_stats_day_index ON country_stats(day);
