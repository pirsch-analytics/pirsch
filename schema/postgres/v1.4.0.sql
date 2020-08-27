ALTER TABLE "hit" ADD COLUMN "os" character varying(20);
ALTER TABLE "hit" ADD COLUMN "os_version" character varying(20);
ALTER TABLE "hit" ADD COLUMN "browser" character varying(20);
ALTER TABLE "hit" ADD COLUMN "browser_version" character varying(20);
ALTER TABLE "hit" ADD COLUMN "desktop" boolean DEFAULT FALSE;
ALTER TABLE "hit" ADD COLUMN "mobile" boolean DEFAULT FALSE;

CREATE TABLE "visitors_per_os" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    os character varying(20),
    os_version character varying(20),
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_os_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_os_id_seq OWNED BY "visitors_per_os".id;
ALTER TABLE ONLY "visitors_per_os" ALTER COLUMN id SET DEFAULT nextval('visitors_per_os_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_os" ADD CONSTRAINT visitors_per_os_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_os_tenant_id_index ON visitors_per_os(tenant_id);
CREATE INDEX visitors_per_os_day_index ON visitors_per_os(day);

CREATE TABLE "visitors_per_browser" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    browser character varying(20),
    browser_version character varying(20),
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_browser_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_browser_id_seq OWNED BY "visitors_per_browser".id;
ALTER TABLE ONLY "visitors_per_browser" ALTER COLUMN id SET DEFAULT nextval('visitors_per_browser_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_browser" ADD CONSTRAINT visitors_per_browser_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_browser_tenant_id_index ON visitors_per_browser(tenant_id);
CREATE INDEX visitors_per_browser_day_index ON visitors_per_browser(day);

CREATE TABLE "visitor_platform" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    desktop integer NOT NULL,
    mobile integer NOT NULL,
    unknown integer NOT NULL
);

CREATE SEQUENCE visitor_platform_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitor_platform_id_seq OWNED BY "visitor_platform".id;
ALTER TABLE ONLY "visitor_platform" ALTER COLUMN id SET DEFAULT nextval('visitor_platform_id_seq'::regclass);
ALTER TABLE ONLY "visitor_platform" ADD CONSTRAINT visitor_platform_pkey PRIMARY KEY (id);
CREATE INDEX visitor_platform_tenant_id_index ON visitor_platform(tenant_id);
CREATE INDEX visitor_platform_day_index ON visitor_platform(day);
