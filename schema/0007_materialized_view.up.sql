ALTER TABLE event DROP COLUMN entry_path;
ALTER TABLE event DROP COLUMN page_views;
ALTER TABLE event DROP COLUMN is_bounce;

CREATE MATERIALIZED VIEW events
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(time)
ORDER BY (time, fingerprint, session_id, path)
POPULATE AS
    SELECT client_id, fingerprint, session_id, event_name, event_meta_keys, event_meta_values, path, title, time,
    duration_seconds, language, country_code, city, referrer, referrer_name, referrer_icon,
    os, os_version, browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term
    FROM (
        SELECT client_id,
        fingerprint,
        session_id,
        event_name,
        event_meta_keys,
        event_meta_values,
        arrayJoin(path_title) path_title,
        tupleElement(path_title, 1) path,
        tupleElement(path_title, 2) title,
        time,
        duration_seconds,
        language,
        country_code,
        city,
        referrer,
        referrer_name,
        referrer_icon,
        os,
        os_version,
        browser,
        browser_version,
        desktop,
        mobile,
        screen_width,
        screen_height,
        screen_class,
        utm_source,
        utm_medium,
        utm_campaign,
        utm_content,
        utm_term
        FROM (
            SELECT client_id,
            fingerprint,
            session_id,
            event_name,
            event_meta_keys,
            event_meta_values,
            groupArray(tuple(path, title)) path_title,
            time,
            sum (duration_seconds) duration_seconds,
            argMax(language, time) language,
            argMax(country_code, time) country_code,
            argMax(city, time) city,
            argMax(referrer, time) referrer,
            argMax(referrer_name, time) referrer_name,
            argMax(referrer_icon, time) referrer_icon,
            argMax(os, time) os,
            argMax(os_version, time) os_version,
            argMax(browser, time) browser,
            argMax(browser_version, time) browser_version,
            argMax(desktop, time) desktop,
            argMax(mobile, time) mobile,
            argMax(screen_width, time) screen_width,
            argMax(screen_height, time) screen_height,
            argMax(screen_class, time) screen_class,
            argMax(utm_source, time) utm_source,
            argMax(utm_medium, time) utm_medium,
            argMax(utm_campaign, time) utm_campaign,
            argMax(utm_content, time) utm_content,
            argMax(utm_term, time) utm_term
            FROM (
                SELECT * FROM event ORDER BY client_id, fingerprint, session_id, time
            )
            GROUP BY client_id, fingerprint, session_id, time, event_name, event_meta_keys, event_meta_values
        )
        GROUP BY client_id, fingerprint, session_id, event_name, event_meta_keys, event_meta_values, path_title, path, title, time,
        duration_seconds, language, country_code, city, referrer, referrer_name, referrer_icon,
        os, os_version, browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
        utm_source, utm_medium, utm_campaign, utm_content, utm_term
    )
;
