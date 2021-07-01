create or replace view operation_groups as
	select network, hash, max(status) as status, max(source) as source, max(expiration_level) as expiration_level, max(level) as level, max(branch) as branch, sum(fee) as fee, max(counter) as max_counter, sum(storage_limit) as storage_limit, sum(gas_limit) as gas_limit, count(*) as num_contents, min(created_at) as created_at from
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