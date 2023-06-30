CREATE OR REPLACE VIEW dipdup_head_status AS
SELECT
    index_name,
    CASE
        WHEN timestamp < NOW() - interval '3 minutes' THEN 'OUTDATED'
        ELSE 'OK'
    END AS status,
    created_at,
    updated_at
FROM dipdup_state;

comment on column dipdup_head_status.index_name is 'Name of the index.';
comment on column dipdup_head_status.status is 'Status of head ("OK" or "OUTDATED" if relevance of head is more than three minutes behind)';
comment on column dipdup_head_status.created_at is 'Date of creation in seconds since UNIX epoch.';
comment on column dipdup_head_status.updated_at is 'Date of last update in seconds since UNIX epoch.';
