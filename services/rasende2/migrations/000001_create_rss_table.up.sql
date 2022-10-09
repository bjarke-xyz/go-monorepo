create table if not exists rss_items(
    item_id text primary key,
    site_name text not null,
    title text not null,
    content text,
    link text,
    published timestamp
);