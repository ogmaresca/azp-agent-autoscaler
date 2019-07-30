#Example Prometheus/OpenMetrics metrics from AZP Agent Autoscaler

```
# HELP azp_agent_autoscaler_active_agents_count The number of active agents
# TYPE azp_agent_autoscaler_active_agents_count gauge
azp_agent_autoscaler_active_agents_count 0
# HELP azp_agent_autoscaler_azd_call_429_count Counts of Azure Devops calls returning HTTP 429 (Too Many Requests)
# TYPE azp_agent_autoscaler_azd_call_429_count counter
azp_agent_autoscaler_azd_call_429_count 0
# HELP azp_agent_autoscaler_azd_call_count Counts of Azure Devops calls
# TYPE azp_agent_autoscaler_azd_call_count counter
azp_agent_autoscaler_azd_call_count{operation="ListJobRequests"} 52
azp_agent_autoscaler_azd_call_count{operation="ListPoolAgents"} 52
azp_agent_autoscaler_azd_call_count{operation="ListPools"} 1
# HELP azp_agent_autoscaler_azd_call_duration_seconds Duration of Azure Devops calls
# TYPE azp_agent_autoscaler_azd_call_duration_seconds histogram
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.005"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.01"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.025"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.05"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.1"} 41
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.25"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="0.5"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="1"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="2.5"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="5"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="10"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListJobRequests",le="+Inf"} 52
azp_agent_autoscaler_azd_call_duration_seconds_sum{operation="ListJobRequests"} 4.910486343000001
azp_agent_autoscaler_azd_call_duration_seconds_count{operation="ListJobRequests"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.005"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.01"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.025"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.05"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.1"} 46
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.25"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="0.5"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="1"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="2.5"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="5"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="10"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPoolAgents",le="+Inf"} 52
azp_agent_autoscaler_azd_call_duration_seconds_sum{operation="ListPoolAgents"} 4.054395294000001
azp_agent_autoscaler_azd_call_duration_seconds_count{operation="ListPoolAgents"} 52
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.005"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.01"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.025"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.05"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.1"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.25"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="0.5"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="1"} 0
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="2.5"} 1
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="5"} 1
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="10"} 1
azp_agent_autoscaler_azd_call_duration_seconds_bucket{operation="ListPools",le="+Inf"} 1
azp_agent_autoscaler_azd_call_duration_seconds_sum{operation="ListPools"} 1.87761948
azp_agent_autoscaler_azd_call_duration_seconds_count{operation="ListPools"} 1
# HELP azp_agent_autoscaler_failed_agents_count The number of failed agents
# TYPE azp_agent_autoscaler_failed_agents_count gauge
azp_agent_autoscaler_failed_agents_count 0
# HELP azp_agent_autoscaler_liveness_probe_count The total number of liveness probes
# TYPE azp_agent_autoscaler_liveness_probe_count counter
azp_agent_autoscaler_liveness_probe_count 52
# HELP azp_agent_autoscaler_pending_agents_count The number of pending agents
# TYPE azp_agent_autoscaler_pending_agents_count gauge
azp_agent_autoscaler_pending_agents_count 1
# HELP azp_agent_autoscaler_queued_pods_count The number of queued pods
# TYPE azp_agent_autoscaler_queued_pods_count gauge
azp_agent_autoscaler_queued_pods_count 0
# HELP azp_agent_autoscaler_scale_down_count The total number of scale downs
# TYPE azp_agent_autoscaler_scale_down_count counter
azp_agent_autoscaler_scale_down_count 0
# HELP azp_agent_autoscaler_scale_down_limited_count The total number of scale downs prevented due to limits
# TYPE azp_agent_autoscaler_scale_down_limited_count counter
azp_agent_autoscaler_scale_down_limited_count 0
# HELP azp_agent_autoscaler_scale_size The size of the agent scaling
# TYPE azp_agent_autoscaler_scale_size gauge
azp_agent_autoscaler_scale_size 0
# HELP azp_agent_autoscaler_scale_up_count The total number of scale ups
# TYPE azp_agent_autoscaler_scale_up_count counter
azp_agent_autoscaler_scale_up_count 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 2.0528e-05
go_gc_duration_seconds{quantile="0.25"} 2.5116e-05
go_gc_duration_seconds{quantile="0.5"} 3.6207e-05
go_gc_duration_seconds{quantile="0.75"} 0.032309435
go_gc_duration_seconds{quantile="1"} 0.099431595
go_gc_duration_seconds_sum 0.397208914
go_gc_duration_seconds_count 23
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 11
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.12.7"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 4.193728e+06
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 4.8647544e+07
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.45831e+06
# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 501572
# HELP go_memstats_gc_cpu_fraction The fraction of this program's available CPU time used by the GC since the program started.
# TYPE go_memstats_gc_cpu_fraction gauge
go_memstats_gc_cpu_fraction 0.0008766049133757522
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 2.38592e+06
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 4.193728e+06
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 5.894144e+07
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 7.217152e+06
# HELP go_memstats_heap_objects Number of allocated objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 37688
# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS.
# TYPE go_memstats_heap_released_bytes gauge
go_memstats_heap_released_bytes 0
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 6.6158592e+07
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 1.5644515417088141e+09
# HELP go_memstats_lookups_total Total number of pointer lookups.
# TYPE go_memstats_lookups_total counter
go_memstats_lookups_total 0
# HELP go_memstats_mallocs_total Total number of mallocs.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 539260
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 27776
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 32768
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 103248
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 114688
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 5.692992e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 1.44805e+06
# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 950272
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 950272
# HELP go_memstats_sys_bytes Number of bytes obtained from system.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 7.25486e+07
# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads 14
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0.6
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 65536
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 9
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 2.6443776e+07
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.56445104338e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 1.340416e+08
# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge
process_virtual_memory_max_bytes -1
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 1
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
```