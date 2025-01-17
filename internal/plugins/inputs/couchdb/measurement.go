// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package couchdb

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info ...
// See also: all *.cfg file in https://github.com/apache/couchdb
//
//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "couchdb",
		Type: "metric",
		Fields: map[string]interface{}{
			"auth_cache_hits_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of authentication cache hits."},
			"auth_cache_misses_total":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of authentication cache misses."},
			"auth_cache_requests_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of authentication cache requests."},
			"coalesced_updates_interactive":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of coalesced interactive updates."},
			"coalesced_updates_replicated":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of coalesced replicated updates."},
			"collect_results_time_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Microsecond latency for calls to couch_db:collect_results."},
			"commits_total":                                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of commits performed."},
			"couch_log_requests_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of logged `level` messages. `level` = `alert` `critical` `debug` `emergency` `error` `info` `notice` `warning`."},
			"couch_replicator_changes_manager_deaths_total":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed replicator changes managers."},
			"couch_replicator_changes_queue_deaths_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed replicator changes work queues."},
			"couch_replicator_changes_read_failures_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed replicator changes read failures."},
			"couch_replicator_changes_reader_deaths_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed replicator changes readers."},
			"couch_replicator_checkpoints_failure_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed checkpoint saves."},
			"couch_replicator_checkpoints_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of checkpoints successfully saves."},
			"couch_replicator_cluster_is_stable":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "1 if cluster is stable, 0 if unstable."},
			"couch_replicator_connection_acquires_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times connections are shared."},
			"couch_replicator_connection_closes_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times a worker is gracefully shut down."},
			"couch_replicator_connection_creates_total":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of connections created."},
			"couch_replicator_connection_owner_crashes_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times a connection owner crashes while owning at least one connection."},
			"couch_replicator_connection_releases_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times ownership of a connection is released."},
			"couch_replicator_connection_worker_crashes_total":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times a worker unexpectedly terminates."},
			"couch_replicator_db_scans_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times replicator db scans have been started."},
			"couch_replicator_docs_completed_state_updates_total": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `completed` state document updates."},
			"couch_replicator_docs_db_changes_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of db changes processed by replicator doc processor."},
			"couch_replicator_docs_dbs_created_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of db shard creations seen by replicator doc processor."},
			"couch_replicator_docs_dbs_deleted_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of db shard deletions seen by replicator doc processor."},
			"couch_replicator_docs_dbs_found_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of db shard found by replicator doc processor."},
			"couch_replicator_docs_failed_state_updates_total":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `failed` state document updates."},
			"couch_replicator_failed_starts_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of replications that have failed to start."},
			"couch_replicator_jobs_adds_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of jobs added to replicator scheduler."},
			"couch_replicator_jobs_crashed":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Replicator scheduler crashed jobs."},
			"couch_replicator_jobs_crashes_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of job crashed noticed by replicator scheduler."},
			"couch_replicator_jobs_duplicate_adds_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of duplicate jobs added to replicator scheduler."},
			"couch_replicator_jobs_pending":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Replicator scheduler pending jobs."},
			"couch_replicator_jobs_removes_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of jobs removed from replicator scheduler."},
			"couch_replicator_jobs_running":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Replicator scheduler running jobs."},
			"couch_replicator_jobs_starts_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of jobs started by replicator scheduler."},
			"couch_replicator_jobs_stops_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of jobs stopped by replicator scheduler."},
			"couch_replicator_jobs_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of replicator scheduler jobs."},
			"couch_replicator_requests_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP requests made by the replicator."},
			"couch_replicator_responses_failure_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed HTTP responses received by the replicator."},
			"couch_replicator_responses_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of successful HTTP responses received by the replicator."},
			"couch_replicator_stream_responses_failure_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed streaming HTTP responses received by the replicator."},
			"couch_replicator_stream_responses_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of successful streaming HTTP responses received by the replicator."},
			"couch_replicator_worker_deaths_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed replicator workers."},
			"couch_replicator_workers_started_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of replicator workers started."},
			"couch_server_lru_skip_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of couch_server LRU operations skipped."},
			"database_purges_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times a database was purged."},
			"database_reads_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times a document was read from a database."},
			"database_writes_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times a database was changed."},
			"db_open_time_seconds":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Milliseconds required to open a database."},
			"dbinfo_seconds":                                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Milliseconds required to DB info."},
			"ddoc_cache_hit_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of design doc cache hits."},
			"ddoc_cache_miss_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of design doc cache misses."},
			"ddoc_cache_recovery_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of design doc cache recoveries."},
			"ddoc_cache_requests_failures_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of design doc cache requests failures."},
			"ddoc_cache_requests_recovery_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of design doc cache requests recoveries."},
			"ddoc_cache_requests_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of design doc cache requests."},
			"document_inserts_total":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of documents inserted."},
			"document_purges_failure_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed document purge operations."},
			"document_purges_success_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of successful document purge operations."},
			"document_purges_total_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of total document purge operations."},
			"document_writes_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of document write operations."},
			"dreyfus_httpd_search_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Distribution of overall search request latency as experienced by the end user."},
			"dreyfus_index_await_seconds":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of an dreyfus_index await request."},
			"dreyfus_index_group1_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of an dreyfus_index group1 request."},
			"dreyfus_index_group2_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of an dreyfus_index group2 request."},
			"dreyfus_index_info_seconds":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of an dreyfus_index info request."},
			"dreyfus_index_search_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of an dreyfus_index search request."},
			"dreyfus_rpc_group1_seconds":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of a group1 RPC worker."},
			"dreyfus_rpc_group2_seconds":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of a group2 RPC worker."},
			"dreyfus_rpc_info_seconds":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of an info RPC worker."},
			"dreyfus_rpc_search_seconds":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of a search RPC worker."},
			"erlang_context_switches_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of context switches."},
			"erlang_dirty_cpu_scheduler_queues":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of all dirty CPU scheduler run queues."},
			"erlang_distribution_recv_avg_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average size of packets, in bytes, received by the socket."},
			"erlang_distribution_recv_cnt_packets_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of packets received by the socket."},
			"erlang_distribution_recv_dvi_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average packet size deviation, in bytes, received by the socket."},
			"erlang_distribution_recv_max_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of the largest packet, in bytes, received by the socket."},
			"erlang_distribution_recv_oct_bytes_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Number of bytes received by the socket."},
			"erlang_distribution_send_avg_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average size of packets, in bytes, sent by the socket."},
			"erlang_distribution_send_cnt_packets_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of packets sent by the socket."},
			"erlang_distribution_send_max_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of the largest packet, in bytes, sent by the socket."},
			"erlang_distribution_send_oct_bytes_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Number of bytes sent by the socket."},
			"erlang_distribution_send_pend_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes waiting to be sent by the socket."},
			"erlang_ets_table":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of ETS tables."},
			"erlang_gc_collections_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of garbage collections by the Erlang emulator."},
			"erlang_gc_words_reclaimed_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of words reclaimed by garbage collections."},
			"erlang_io_recv_bytes_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The total number of bytes received through ports."},
			"erlang_io_sent_bytes_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The total number of bytes output to ports."},
			"erlang_memory_bytes":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of memory (in bytes) dynamically allocated by the Erlang emulator."},
			"erlang_message_queue_max":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Maximum size across all message queues."},
			"erlang_message_queue_min":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Minimum size across all message queues."},
			"erlang_message_queue_size":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Size of message queue."},
			"erlang_message_queues":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total size of all message queues."},
			"erlang_process_limit":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The maximum number of simultaneously existing Erlang processes."},
			"erlang_processes":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of Erlang processes."},
			"erlang_reductions_total":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of reductions."},
			"erlang_scheduler_queues":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of all normal run queues."},
			"fabric_doc_update_errors_total":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of document update errors."},
			"fabric_doc_update_mismatched_errors_total":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of document update errors with multiple error types."},
			"fabric_doc_update_write_quorum_errors_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of write quorum errors."},
			"fabric_open_shard_timeouts_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of open shard timeouts."},
			"fabric_read_repairs_failures_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed read repair operations."},
			"fabric_read_repairs_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of successful read repair operations."},
			"fabric_worker_timeouts_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of worker timeouts."},
			"fsync_count_total":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of fsync calls."},
			"fsync_time":                                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Microseconds to call fsync."},
			"global_changes_db_writes_total":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of db writes performed by global changes."},
			"global_changes_event_doc_conflict_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of conflicted event docs encountered by global changes."},
			"global_changes_listener_pending_updates":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of global changes updates pending writes in global_changes_listener."},
			"global_changes_rpcs_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of rpc operations performed by global_changes."},
			"global_changes_server_pending_updates":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of global changes updates pending writes in global_changes_server."},
			"httpd_aborted_requests_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of aborted requests."},
			"httpd_all_docs_timeouts_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP all_docs timeouts."},
			"httpd_bulk_docs_seconds":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Distribution of the number of docs in _bulk_docs requests."},
			"httpd_bulk_requests_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bulk requests."},
			"httpd_clients_requesting_changes_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of clients for continuous _changes."},
			"httpd_dbinfo":                                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Distribution of latencies for calls to retrieve DB info."},
			"httpd_explain_timeouts_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP _explain timeouts."},
			"httpd_find_timeouts_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP find timeouts."},
			"httpd_partition_all_docs_requests_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP _all_docs requests."},
			"httpd_partition_all_docs_timeouts_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP all_docs timeouts."},
			"httpd_partition_explain_requests_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP _explain requests."},
			"httpd_partition_explain_timeouts_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP _explain timeouts."},
			"httpd_partition_find_requests_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP _find requests."},
			"httpd_partition_find_timeouts_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP find timeouts."},
			"httpd_partition_view_requests_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP view requests."},
			"httpd_partition_view_timeouts_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of partition HTTP view timeouts."},
			"httpd_purge_requests_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of purge requests."},
			"httpd_request_methods":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP `option` requests. `option` = `COPY` `DELETE` `GET` `HEAD` `OPTIONS` `POST` `PUT`."},
			"httpd_requests_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP requests."},
			"httpd_status_codes":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP `status_codes` responses. `status_codes` = 200 201 202 204 206 301 304 400 403 404 405 406 409 412 414 415 416 417 500 501 503."},
			"httpd_temporary_view_reads_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of temporary view reads."},
			"httpd_view_reads_total":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of view reads."},
			"httpd_view_timeouts_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of HTTP view timeouts."},
			"io_queue2_search_count_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Search IO directly triggered by client requests."},
			"io_queue_search_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Search IO directly triggered by client requests."},
			"legacy_checksums":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of legacy checksums found in couch_file instances."},
			"local_document_writes_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of document write operations."},
			"mango_docs_examined_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of documents examined by mango queries coordinated by this node."},
			"mango_evaluate_selector_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of mango selector evaluations."},
			"mango_keys_examined_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of keys examined by mango queries coordinated by this node."},
			"mango_query_invalid_index_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of mango queries that generated an invalid index warning."},
			"mango_query_time_seconds":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Length of time processing a mango query."},
			"mango_quorum_docs_examined_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of documents examined by mango queries, using cluster quorum."},
			"mango_results_returned_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of rows returned by mango queries."},
			"mango_too_many_docs_scanned_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of mango queries that generated an index scan warning."},
			"mango_unindexed_queries_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of mango queries that could not use an index."},
			"mem3_shard_cache_eviction_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of shard cache evictions."},
			"mem3_shard_cache_hit_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of shard cache hits."},
			"mem3_shard_cache_miss_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of shard cache misses."},
			"mrview_emits_total":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of invocations of `emit` in map functions in the view server."},
			"mrview_map_doc_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of documents mapped in the view server."},
			"nouveau_active_searches_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of active search requests."},
			"nouveau_search_latency":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Distribution of overall search request latency as experienced by the end user."},
			"open_databases_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of open databases."},
			"open_os_files_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of file descriptors CouchDB has open."},
			"pread_exceed_eof_total":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of the attempts to read beyond end of db file."},
			"pread_exceed_limit_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of the attempts to read beyond set limit."},
			"query_server_acquired_processes_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of acquired external processes."},
			"query_server_process_errors_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of OS error process exits."},
			"query_server_process_exists_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of OS normal process exits."},
			"query_server_process_prompt_errors_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of OS process prompt errors."},
			"query_server_process_prompts_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of successful OS process prompts."},
			"query_server_process_starts_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of OS process starts."},
			"query_server_vdu_process_time_seconds":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Duration of validate_doc_update function calls."},
			"query_server_vdu_rejects_total":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of rejections by validate_doc_update function."},
			"request_time_seconds":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationDay, Desc: "Length of a request inside CouchDB without `MochiWeb`."},
			"rexi_buffered_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `rexi` messages buffered."},
			"rexi_down_total":                                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `rexi_DOWN` messages handled."},
			"rexi_dropped_total":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `rexi` messages dropped from buffers."},
			"rexi_streams_timeout_stream_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `rexi` stream timeouts."},
			"rexi_streams_timeout_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `rexi` stream initialization timeouts."},
			"rexi_streams_timeout_wait_for_ack_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of `rexi` stream timeouts while waiting for `acks`."},
			"uptime_seconds":                                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "CouchDB uptime."},
		},
		Tags: map[string]interface{}{
			"host":        inputs.NewTagInfo("Host name."),
			"instance":    inputs.NewTagInfo("Instance endpoint."),
			"level":       inputs.NewTagInfo("Log lever, in `alert` `critical` `debug` `emergency` `error` `info` `notice` `warning`."),
			"quantile":    inputs.NewTagInfo("Histogram `quantile`."),
			"method":      inputs.NewTagInfo("HTTP requests type, in `COPY` `DELETE` `GET` `HEAD` `OPTIONS` `POST` `PUT`."),
			"code":        inputs.NewTagInfo("Code of HTTP responses, in 200 201 202 204 206 301 304 400 403 404 405 406 409 412 414 415 416 417 500 501 503."),
			"memory_type": inputs.NewTagInfo("Erlang memory type, in `total` `processes` `processes_used` `system` `atom` `atom_used` `binary` `code` `ets`"),
			"stage":       inputs.NewTagInfo("`Rexi` stream stage, like `init_stream`."),
		},
	}
}
