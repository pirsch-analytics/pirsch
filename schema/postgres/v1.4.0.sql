ALTER TABLE "hit" ADD COLUMN "os" character varying(20);
ALTER TABLE "hit" ADD COLUMN "os_version" character varying(20);
ALTER TABLE "hit" ADD COLUMN "browser" character varying(20);
ALTER TABLE "hit" ADD COLUMN "browser_version" character varying(20);

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
