create or replace view mutez_per_gas_unit as
	select gas.waiting_levels,
	       max(mutez_per_gas_unit),
	       min(mutez_per_gas_unit),
	       avg(mutez_per_gas_unit) as avg,
	       count(mutez_per_gas_unit),
	       percentile_disc(0.5) within group (order by gas.mutez_per_gas_unit) as median
        from (
		    select
                (level_in_chain - level_in_mempool) as waiting_levels,
                ((total_fee - 100 - 150 * 1)::float / total_gas_used) as mutez_per_gas_unit
                from gas_stats gs
                    where level_in_chain > 0
                        and level_in_mempool > 0
                        and total_gas_used > 0
        ) as gas
        group by gas.waiting_levels;

comment on view mutez_per_gas_unit is 'Statistics for micro tez per gas unit.';
comment on column mutez_per_gas_unit.waiting_levels is 'Difference between level in chain and level in mempool.';
comment on column mutez_per_gas_unit.max is 'Maximum price for gas unit.';
comment on column mutez_per_gas_unit.min is 'Minimum price for gas unit.';
comment on column mutez_per_gas_unit.avg is 'Average price for gas unit.';
comment on column mutez_per_gas_unit.count is 'Count of prices for gas unit.';
comment on column mutez_per_gas_unit.median is 'Percentile (50%) of price for gas unit.';
