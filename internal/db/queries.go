package db

var MigrateSQL = `
CREATE TABLE IF NOT EXISTS raw_request(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  sent_at timestamp default current_timestamp,
  body  JSON
);
CREATE TABLE IF NOT EXISTS raw_response(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  received_at timestamp default current_timestamp,
  body  JSON
);

CREATE VIEW IF NOT EXISTS response (
  id,
  version,
  type_url,
  resource
)as
  select raw_response.id,
		 json_extract(body, '$.version_info'),
		 json_extract(body, '$.type_url'),
         case json_extract(body, '$.type_url')
	       when 'type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment'
		     then json_extract(value, '$.cluster_name')
		   else
		     json_extract(value, '$.name')
         end
	from raw_response,
		 json_each(body,'$.resources');
`

var InsertRequestSQL = `
insert into raw_request(body)
			values(?);
`
var InsertResponseSQL = `
insert into raw_response(body)
			values(?);
`

var DeleteAllSQL = `
begin;
delete from raw_request;
delete from raw_response;
commit;
`

var CheckExpectedResourcesSQL = `
with expected as (
  select value as resource
	from json_each($1)
), actual as (
  select resource, id
	from response
   where version = $2
	 and type_url = $3
	 and resource in (select resource from expected)
)
	select ((select count(*) from expected) = (select count(*) from actual)),
		   (count(distinct id) = 1)
	  from actual;
`

var CheckMoreRequestsThanResponseSQL = `
select (select count(*) from raw_request) > (select count(*) from raw_response);
`
