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
  resource,
  removed_resource
)as
  select raw.id,
		 case
           when json_extract(body, '$.version_info') is null
             then json_extract(body, '$.system_version_info')
           else
             json_extract(body,'$.version_info')
         end,
		 json_extract(body, '$.type_url'),
         case json_extract(body, '$.type_url')
	       when 'type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment'
		     then json_extract(a.value, '$.cluster_name')
		   else
		     json_extract(a.value, '$.name')
         end,
         x.value
	from raw_response raw
		 left join json_each(body,'$.resources') a
         left join json_each(body, '$.removed_resources') x;
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

var CheckOnlyExpectedResourcesSQL = `
with expected as (
  select value as resource
	from json_each($1)
), match_version as (
  select distinct version
    from response
   where version  = $2
     and type_url = $3
     and resource in (select resource from expected)
), all_for_version as(
  select *
    from response
    join match_version on response.version = match_version.version
)
select ((select count(*) from expected) = (select count(*) from all_for_version));
`

var DeltaCheckOnlyExpectedResourcesSQL = `
with expected as (
  select value as resource
    from json_each($1)
), latest_match as (
  select resource
    from response
   where version = ($2)
     and type_url = ($3)
     and id = (select max(id) from response)
)
select ((select count(*) from expected) = (select count(*) from latest_match));
`

var DeltaCheckRemovedResourcesSql = `
  with expected as (
    select value as resource
      from json_each($1)
), latest_match as (
    select removed_resource
      from response
     where type_url = ($2)
       and id = (select max(id) from raw_response)
)

select ((select count(*) from expected) = (select count(*) from latest_match));
`

var DeltaCheckNoResourceSql = `
select count(*) = 0
  from response
  where resource = ($1)
    and version = ($2)
    and type_url = ($3);
`

var CheckMoreRequestsThanResponseSQL = `
select (select count(*) from raw_request) > (select count(*) from raw_response);
`

var CheckNoResponsesForVersionSQL = `
select (count(*) = 0)
  from response
 where version = $1;
`
