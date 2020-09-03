ALTER TABLE "hit" RENAME COLUMN "ref" TO "referrer";

CREATE TABLE "visitor_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    visitors integer NOT NULL,
    platform_desktop integer NOT NULL,
    platform_mobile integer NOT NULL,
    platform_unknown integer NOT NULL
);

CREATE SEQUENCE visitor_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitor_stats_id_seq OWNED BY "visitor_stats".id;
ALTER TABLE ONLY "visitor_stats" ALTER COLUMN id SET DEFAULT nextval('visitor_stats_id_seq'::regclass);
ALTER TABLE ONLY "visitor_stats" ADD CONSTRAINT visitor_stats_pkey PRIMARY KEY (id);
CREATE INDEX visitor_stats_day_index ON visitor_stats(day);
CREATE INDEX visitor_stats_path_index ON visitor_stats(path);

CREATE TABLE "visitor_time_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    hour smallint NOT NULL,
    visitors integer NOT NULL
);

CREATE SEQUENCE visitor_time_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitor_time_stats_id_seq OWNED BY "visitor_time_stats".id;
ALTER TABLE ONLY "visitor_time_stats" ALTER COLUMN id SET DEFAULT nextval('visitor_time_stats_id_seq'::regclass);
ALTER TABLE ONLY "visitor_time_stats" ADD CONSTRAINT visitor_time_stats_pkey PRIMARY KEY (id);
CREATE INDEX visitor_time_stats_day_index ON visitor_time_stats(day);
CREATE INDEX visitor_time_stats_path_index ON visitor_time_stats(path);

CREATE TABLE "language_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    language varchar(10),
    visitors integer NOT NULL
);

CREATE SEQUENCE language_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE language_stats_id_seq OWNED BY "language_stats".id;
ALTER TABLE ONLY "language_stats" ALTER COLUMN id SET DEFAULT nextval('language_stats_id_seq'::regclass);
ALTER TABLE ONLY "language_stats" ADD CONSTRAINT language_stats_pkey PRIMARY KEY (id);
CREATE INDEX language_stats_day_index ON language_stats(day);
CREATE INDEX language_stats_path_index ON language_stats(path);

CREATE TABLE "referrer_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    referrer varchar(2000),
    visitors integer NOT NULL
);

CREATE SEQUENCE referrer_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE referrer_stats_id_seq OWNED BY "referrer_stats".id;
ALTER TABLE ONLY "referrer_stats" ALTER COLUMN id SET DEFAULT nextval('referrer_stats_id_seq'::regclass);
ALTER TABLE ONLY "referrer_stats" ADD CONSTRAINT referrer_stats_pkey PRIMARY KEY (id);
CREATE INDEX referrer_stats_day_index ON referrer_stats(day);
CREATE INDEX referrer_stats_path_index ON referrer_stats(path);

CREATE TABLE "os_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    os character varying(20),
    os_version character varying(20),
    visitors integer NOT NULL
);

CREATE SEQUENCE os_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE os_stats_id_seq OWNED BY "os_stats".id;
ALTER TABLE ONLY "os_stats" ALTER COLUMN id SET DEFAULT nextval('os_stats_id_seq'::regclass);
ALTER TABLE ONLY "os_stats" ADD CONSTRAINT os_stats_pkey PRIMARY KEY (id);
CREATE INDEX os_stats_day_index ON os_stats(day);
CREATE INDEX os_stats_path_index ON os_stats(path);

CREATE TABLE "browser_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    browser character varying(20),
    browser_version character varying(20),
    visitors integer NOT NULL
);

CREATE SEQUENCE browser_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE browser_stats_id_seq OWNED BY "browser_stats".id;
ALTER TABLE ONLY "browser_stats" ALTER COLUMN id SET DEFAULT nextval('browser_stats_id_seq'::regclass);
ALTER TABLE ONLY "browser_stats" ADD CONSTRAINT browser_stats_pkey PRIMARY KEY (id);
CREATE INDEX browser_stats_day_index ON browser_stats(day);
CREATE INDEX browser_stats_path_index ON browser_stats(path);

DROP TABLE "visitor_platform";
DROP TABLE "visitors_per_browser";
DROP TABLE "visitors_per_day";
DROP TABLE "visitors_per_hour";
DROP TABLE "visitors_per_language";
DROP TABLE "visitors_per_os";
DROP TABLE "visitors_per_page";
DROP TABLE "visitors_per_referrer";
