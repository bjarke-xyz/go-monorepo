alter table rss_items add column if not exists ts_title tsvector
    generated always as (to_tsvector('danish', title)) stored;
create index if not exists ts_title_idx on rss_items using GIN(ts_title);

alter table rss_items add column if not exists ts_content tsvector
    generated always as (to_tsvector('danish', content)) stored;
create index if not exists ts_content_idx on rss_items using GIN(ts_content)