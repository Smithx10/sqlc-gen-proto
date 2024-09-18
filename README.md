# sqlc-gen-proto Generates .proto files


## Requires a forked sqlc at the moment.
https://github.com/Smithx10/sqlc/tree/comments_to_plugin


## Usage
#### Problem:
When developing middleware that calls into generated code from sqlc we ended up doing a lot of copy and paste.  This is an attempt to create the following iteration workflow.

1. Create Schema and Queries With Plugin Annotations
2. Generate Golang Code, and Protobufs 
3. Create Proto Request_Response and Proto Services using Messages generated from this plugin.
4. Repeat endlessly.



### Current Comment Annotation Options:
#### Annotation Defaults: 
| Name | Default Value |
| -------------- | --------------- |
| "package" | "sqlcgen" |
| "messagename" | tablename |


#### -- generate:
*"-- generate:"*  specifies if the table should be generated.

#### -- package: <name>
*"-- package:"*  specifies the package for the given protobuf file.

#### -- skip: <field>
*"-- skip:"*  can be applied to a single field to indicate you'd like to not include it in the message.  By default all columns in both queries and tables are added. *can be annotated many times above 1 statement*
#### -- request_response: oneof <field> <field> <field>
*"-- request_response: oneof "*  is used for applying a oneof configuration to the get, update, delete messages.
#### -- request_response: req_feild <field>
*"-- request_response: req_field "*  is used for adding an additional field.  Sometimes APIs require a path.  Ex. /api/v1/orgs/{org}/projects/{project}/resources.  You'd want to add req_field twice to additional fields 
#### -- service: <service> <path>
*"-- serivce: path"*  is used for adding an service. The path is used to define what path to use for the google api http rules.

Path: /v1/users :
| method | path | http |
| --------------- | --------------- | --------------- |
| Create | /v1/users | POST |
| Get | /v1/users/{$primarykey} | GET |
| Update | /v1/users/{$primarykey} | PUT |
| Delete | /v1/users/{$primarykey} | DELETE |
| List | /v1/users | GET |

#### SQL Plugin Config
```json
{
  "version": "2",
  "plugins": [
    {
      "name": "proto",
      "process": {
        "cmd": "sqlc-gen-proto"
      }
    }
  ],
  "sql": [
    {
      "engine": "postgresql",
      "queries": "query.sql",
      "schema": "schema.sql",
      "codegen": [
        {
          "out": "gen",
          "plugin": "proto",
          "options": {
            "out_dir": "./gen",
            "user_defined_dir": "./user_defined",
            "one_of_id": "ident",
            "default_package": "bob"
          }
        }
      ]
    }
  ]
}
```




#### Full Example:
```sql
-- generate:
-- package: foo.bar.baz.v1
-- skip: alias 
-- skip: name
-- request_response: oneof uuid name
-- request_response: req_field string project 
-- service: IAM /v1/users
CREATE TABLE "public"."users" (
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  "name" character varying NOT NULL,
  "alias" character varying NULL
);
```



#### Type Conversion Default
| Postgres | Not Null Proto | Null Proto |
| --------------- | --------------- | --------------- |
| integer | int32 | Int32Value |
| int | int32 | Int32Value |
| int4 | int32 | Int32Value |
| pg_catalog.int4 | int32 | Int32Value |
| serial | int32 | Int32Value |
| serial4 | int32 | Int32Value |
| pg_catalog.serial4 | int32 | Int32Value |
| smallserial | int32 | Int32Value |
| smallint | int32 | Int32Value |
| int2 | int32 | Int32Value |
| pg_catalog.int2" | int32 | Int32Value |
| serial2 | int32 | Int32Value |
| pg_catalog.serial2 | int32 | Int32Value |
| interval | int64 | Int64Value |
| pg_catalog.interval | int64 | Int64Value |
| bigint | int64 | Int64Value |
| int8 | int64 | Int64Value |
| pg_catalog.int8 | int64 | Int64Value |
| bigserial | int64 | Int64Value |
| serial8 | int64 | Int64Value |
| pg_catalog.serial8 | int64 | Int64Value |
| real | Float | FloatValue |
| float4 | Float | FloatValue |
| pg_catalog.float4 | Float | FloatValue |
| float | Float | FloatValue |
| double precision | Float | FloatValue |
| float8 | Float | FloatValue |
| pg_catalog.float8 | Float | FloatValue |
| numeric" | Decimal | Decimal |
| pg_catalog.numeric" | Decimal | Decimal |
| money | Money | Money |
| boolean | bool | BoolValue |
| bool | bool | BoolValue |
| pg_catalog.bool | bool | BoolValue |
| json | Struct | Struct |
| uuid | bytes | BytesValue |
| jsonb | bytes | BytesValue |
| bytea | bytes | BytesValue |
| blob | bytes | BytesValue |
| pg_catalog.bytea | bytes | BytesValue |
| pg_catalog.timestamptz | Timestamp | Timestamp |
| date | Timestamp | Timestamp |
| timestamptz | Timestamp | Timestamp |
| pg_catalog.timestamp | Timestamp | Timestamp |
| pg_catalog.timetz | Timestamp | Timestamp |
| pg_catalog.time | Timestamp | Timestamp |
| void | Any | Any |

This Schema Defintion Generates the following directory structure and files.
```sql
-- generate:
-- package: baz.bar.foo.v1
-- request_response: oneof uuid name
-- request_response: req_field string project 
-- service: IAM /v1/users
CREATE TABLE "public"."users" (
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  "name" character varying NOT NULL,
  "alias" character varying NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL DEFAULT NOW(),
  "updated_at" timestamptz NOT NULL DEFAULT NOW(),
  "deleted_at" timestamptz
);

-- generate:
-- package: baz.bar.foo.v1
-- request_response: oneof uuid name
-- request_response: req_field string project 
-- service: IAM /v1/groups
CREATE TABLE "public"."groups" (
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY, 
  "name" character varying NOT NULL,
  "alias" character varying NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL DEFAULT NOW(),
  "updated_at" timestamptz NOT NULL DEFAULT NOW(),
  "deleted_at" timestamptz NULL
);

```

```bash 
gen
├── baz
│   └── bar
│       └── foo
│           └── v1
│               ├── enum.proto
│               ├── message.proto
│               ├── request_response.proto
│               └── service.proto

```

enum.proto
```proto

syntax = "proto3";

package baz.bar.foo.v1;

enum ResourceType {
  RESOURCE_TYPE_UNSPECIFIED = 0;

  RESOURCE_TYPE_ORGANIZATION = 1;

  RESOURCE_TYPE_FOLDER = 2;

  RESOURCE_TYPE_PROJECT = 3;
}
```
message.proto
```proto
syntax = "proto3";

package baz.bar.foo.v1;

import "google/protobuf/timestamp.proto";

import "google/protobuf/wrappers.proto";

message Users {
  bytes uuid = 1;

  string name = 2;

  google.protobuf.StringValue alias = 3;

  google.protobuf.StringValue description = 4;

  google.protobuf.Timestamp created_at = 5;

  google.protobuf.Timestamp updated_at = 6;

  google.protobuf.Timestamp deleted_at = 7;

  repeated string memberof = 8;
}

message Groups {
  bytes uuid = 1;

  string name = 2;

  google.protobuf.StringValue alias = 3;

  google.protobuf.StringValue description = 4;

  google.protobuf.Timestamp created_at = 5;

  google.protobuf.Timestamp updated_at = 6;

  google.protobuf.Timestamp deleted_at = 7;
}

```

request_response.proto
```proto
syntax = "proto3";

package baz.bar.foo.v1;

import "baz/bar/foo/v1/message.proto";

message CreateUsersRequest {
  string project = 1;

  Users users = 2;
}

message CreateUsersResponse {
  Users users = 1;
}

message GetUsersRequest {
  string project = 1;

  oneof ident {
    bytes uuid = 2;

    string name = 3;
  }
}

message GetUsersResponse {
  Users users = 1;
}

message UpdateUsersRequest {
  string project = 1;

  oneof ident {
    bytes uuid = 2;

    string name = 3;
  }

  Users users = 4;
}

message UpdateUsersResponse {
  Users users = 1;
}

message DeleteUsersRequest {
  string project = 1;

  oneof ident {
    bytes uuid = 2;

    string name = 3;
  }
}

message DeleteUsersResponse {
}

message ListUsersRequest {
  string project = 1;

  int32 page_size = 2;

  string page_token = 3;
}

message ListUsersResponse {
  string next_page_token = 1;

  repeated Users users = 2;
}

message CreateGroupsRequest {
  string project = 1;

  Groups groups = 2;
}

message CreateGroupsResponse {
  Groups groups = 1;
}

message GetGroupsRequest {
  string project = 1;

  oneof ident {
    bytes uuid = 2;

    string name = 3;
  }
}

message GetGroupsResponse {
  Groups groups = 1;
}

message UpdateGroupsRequest {
  string project = 1;

  oneof ident {
    bytes uuid = 2;

    string name = 3;
  }

  Groups groups = 4;
}

message UpdateGroupsResponse {
  Groups groups = 1;
}

message DeleteGroupsRequest {
  string project = 1;

  oneof ident {
    bytes uuid = 2;

    string name = 3;
  }
}

message DeleteGroupsResponse {
}

message ListGroupsRequest {
  string project = 1;

  int32 page_size = 2;

  string page_token = 3;
}

message ListGroupsResponse {
  string next_page_token = 1;

  repeated Groups groups = 2;
}
```

service.proto
```proto
syntax = "proto3";

package baz.bar.foo.v1;

import "baz/bar/foo/v1/request_response.proto";

import "google/api/annotations.proto";

service Iam {
  rpc CreateUsers ( CreateUsersRequest ) returns ( CreateUsersResponse ) {
    option (google.api.http) = { post: "/v1/users", body: "*" };
  }

  rpc GetUsers ( GetUsersRequest ) returns ( GetUsersResponse ) {
    option (google.api.http) = { get: "/v1/users/{uuid}" };
  }

  rpc UpdateUsers ( UpdateUsersRequest ) returns ( UpdateUsersResponse ) {
    option (google.api.http) = { put: "/v1/users/{uuid}", body: "*" };
  }

  rpc DeleteUsers ( DeleteUsersRequest ) returns ( DeleteUsersResponse ) {
    option (google.api.http) = { delete: "/v1/users/{uuid}" };
  }

  rpc ListUsers ( ListUsersRequest ) returns ( ListUsersResponse ) {
    option (google.api.http) = { get: "/v1/users" };
  }

  rpc CreateGroups ( CreateGroupsRequest ) returns ( CreateGroupsResponse ) {
    option (google.api.http) = { post: "/v1/groups", body: "*" };
  }

  rpc GetGroups ( GetGroupsRequest ) returns ( GetGroupsResponse ) {
    option (google.api.http) = { get: "/v1/groups/{uuid}" };
  }

  rpc UpdateGroups ( UpdateGroupsRequest ) returns ( UpdateGroupsResponse ) {
    option (google.api.http) = { put: "/v1/groups/{uuid}", body: "*" };
  }

  rpc DeleteGroups ( DeleteGroupsRequest ) returns ( DeleteGroupsResponse ) {
    option (google.api.http) = { delete: "/v1/groups/{uuid}" };
  }

  rpc ListGroups ( ListGroupsRequest ) returns ( ListGroupsResponse ) {
    option (google.api.http) = { get: "/v1/groups" };
  }
}
```

#### Extending Messages, Enums, Request_Response and Services:
enums,messages, and services defined in the user_defined directory that match the generate file path will be appended to the existing generated file.

For Example:
``` sql
-- generate:
-- package: baz.bar.foo.v1
CREATE TABLE "public"."users" (
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  "name" character varying NOT NULL,
  "alias" character varying NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL DEFAULT NOW(),
  "updated_at" timestamptz NOT NULL DEFAULT NOW(),
  "deleted_at" timestamptz
);
```
With the following manually defined...
user_defined/baz/bar/foo/v1/message.proto
``` proto
syntax = "proto3";

package baz.bar.foo.v1;


message Users {
  string test_field = 1;
}

message Groups {
  google.protobuf.StringValue test_field = 1;
}
```

Will result in this message in generated proto
```proto
message Users {
  bytes uuid = 1;

  string name = 2;

  google.protobuf.StringValue alias = 3;

  google.protobuf.StringValue description = 4;

  google.protobuf.Timestamp created_at = 5;

  google.protobuf.Timestamp updated_at = 6;

  google.protobuf.Timestamp deleted_at = 7;

  repeated string memberof = 8;

  string test_field = 9;
}

message Groups {
  bytes uuid = 1;

  string name = 2;

  google.protobuf.StringValue alias = 3;

  google.protobuf.StringValue description = 4;

  google.protobuf.Timestamp created_at = 5;

  google.protobuf.Timestamp updated_at = 6;

  google.protobuf.Timestamp deleted_at = 7;

  google.protobuf.StringValue test_field = 8;
}
```
