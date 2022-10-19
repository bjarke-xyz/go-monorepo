create table if not exists recipes (
    id uuid default gen_random_uuid() primary key,
    title text not null unique,
    description text,
    image_id uuid,
    user_id uuid not null,
    created_at timestamp not null default NOW(),
    moderated_at timestamp,
    last_modified_at timestamp not null default NOW(),
    published boolean not null,
    tips text[],
    yield text,
    parts json
);

create table if not exists files (
    id uuid default gen_random_uuid() primary key,
    bucket text not null,
    key text not null unique,
    content_type text not null,
    size bigint not null,
    name text
);