CREATE table helloworld.ads_data (
        date Date,
        time UInt32,
        event UInt8, // 1,2
        platform UInt8,
        ad_id UUID,
        client_union_id UUID,
        campaign_union_id UUID,
        ad_cost_type UInt8,
        ad_cost UInt16,
        has_video Bool,
        target_audience_count UInt16
) ENGINE = MergeTree() PRIMARY KEY (ad_id, date, time)

// Число event
with lhs AS (select distinct(date) AS dd
FROM helloworld.ads_data)
select dd, cnt from (
    select date, COUNT(*) AS cnt
    FROM helloworld.ads_data
    GROUP BY date) rhs
RIGHT JOIN lhs ON rhs.date = lhs.dd

// Число показов/кликов event = 1/2
with lhs AS (select distinct(date) AS dd
FROM helloworld.ads_data)
select dd, cnt from (
    select date, COUNT(*) AS cnt
    FROM helloworld.ads_data
    WHERE event = 2
    GROUP BY date) rhs
RIGHT JOIN lhs ON rhs.date = lhs.dd

// Число уникальных объявлений
with lhs AS (select distinct(date) AS dd
FROM helloworld.ads_data)
select dd, cnt from (
    select date, COUNT(distinct(ad_id)) AS cnt
    FROM helloworld.ads_data
    GROUP BY date) rhs
RIGHT JOIN lhs ON rhs.date = lhs.dd

// Число уникальных кампаний
with lhs AS (select distinct(date) AS dd
FROM helloworld.ads_data)
select dd, cnt from (
    select date, COUNT(distinct(campaign_union_id)) AS cnt
    FROM helloworld.ads_data
    GROUP BY date) rhs
RIGHT JOIN lhs ON rhs.date = lhs.dd

// Найти объявления по которым показ произошел после клика
with CLICK AS (
    select ad_id, min(toUnixTimestamp(toDateTime(date))+time) AS timestamp
    FROM helloworld.ads_data
    WHERE event = 2
    GROUP BY ad_id
)
select * from (
    select ad_id, min(toUnixTimestamp(toDateTime(date))+time) AS timestamp
    FROM helloworld.ads_data
    WHERE event = 1
    GROUP BY ad_id
) SHOW
JOIN CLICK ON SHOW.ad_id = CLICK.ad_id
HAVING CLICK.timestamp < SHOW.timestamp