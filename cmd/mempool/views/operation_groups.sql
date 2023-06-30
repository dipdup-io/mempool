create or replace view operation_groups as
	select network,
	       hash,
	       max(status) as status,
	       max(source) as source,
	       max(expiration_level) as expiration_level,
	       max(level) as level,
	       max(branch) as branch,
	       sum(fee) as fee,
	       max(counter) as max_counter,
	       sum(storage_limit) as storage_limit,
	       sum(gas_limit) as gas_limit,
	       count(*) as num_contents,
	       min(created_at) as created_at
	from
	    (
            select network, status, source, expiration_level, level, branch, hash, fee, counter, storage_limit, gas_limit, created_at from "transactions"
            union all
            select network, status, source, expiration_level, level, branch, hash, fee, counter, storage_limit, gas_limit, created_at from delegations
            union all
            select network, status, source, expiration_level, level, branch, hash, fee, counter, storage_limit, gas_limit, created_at from originations
            union all
            select network, status, source, expiration_level, level, branch, hash, fee, counter, storage_limit, gas_limit, created_at from reveals
        ) as foo
	group by network, hash;

comment on view operation_groups is 'Statistics per operations (transactions, delegations, originations, reveals) grouped by network and hash.';
comment on column operation_groups.network is 'Network of the group.';
comment on column operation_groups.hash is 'Hash of the operation group.';
comment on column operation_groups.status is 'Status (max) of the operation group.';
comment on column operation_groups.source is 'Source (max) of the operation group.';
comment on column operation_groups.expiration_level is 'Expiration (max) level of the operation group.';
comment on column operation_groups.level is 'Level (max) of the operation group.';
comment on column operation_groups.branch is 'Branch (max) of the operation group.';
comment on column operation_groups.fee is 'Sum of the fee of the operation group.';
comment on column operation_groups.max_counter is 'Maximum counter of the operation group.';
comment on column operation_groups.storage_limit is 'Sum of the storage limit of the operation group.';
comment on column operation_groups.gas_limit is 'Sum of the gas limit of the operation group.';
comment on column operation_groups.num_contents is 'Number of operations in group.';
comment on column operation_groups.created_at is 'Date of fist operation creation in seconds since UNIX epoch.'
