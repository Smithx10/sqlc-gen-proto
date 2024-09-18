# sqlc-gen-proto Generates .proto files


## Requires a forked sqlc at the moment.


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

#### -- package:
*"-- package:"*  specifies the package for the given protobuf file.

#### -- skip:
*"-- skip:"*  can be applied to a single field to indicate you'd like to not include it in the message.  By default all columns in both queries and tables are added. *can be annotated many times above 1 statement*

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

Schema:
```sql
-- generate:
-- package: foo.bar.baz.v1
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
CREATE TABLE "public"."groups" (
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY, 
  "name" character varying NOT NULL,
  "alias" character varying NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL DEFAULT NOW(),
  "updated_at" timestamptz NOT NULL DEFAULT NOW(),
  "deleted_at" timestamptz NULL
);

CREATE TABLE "public"."group_members" (
  "group_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("group_id", "user_id"),
  CONSTRAINT "group_members_group_id" FOREIGN KEY ("group_id") REFERENCES "public"."groups" ("uuid") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "group_members_user_id" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("uuid") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- generate: 
-- package: the.resourcemanager.v1
CREATE TABLE resource (
    resource_uuid UUID PRIMARY KEY UNIQUE,
    resource_name character varying NOT NULL UNIQUE, 
    parent_resource_uuid UUID REFERENCES resource(resource_uuid),
    resource_type resource_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT valid_parent CHECK (resource_type != 'project' OR parent_resource_uuid IS NULL)
);
```


Query:
``` sql
-- name: GetUsers :one
-- generate:
-- package: foo.bar.baz.v1
-- messagename: users
SELECT u.*, COALESCE(array_agg(g.name)::text[], array[]::text[])::text[] as memberof FROM users u left join group_members gm on u.uuid = gm.user_id left join groups g on gm.group_id = g.uuid where u.name = $1 group by u.uuid;
```


Enum: 
```sql
-- generate:
-- package: enum.v1
CREATE TYPE resource_type AS ENUM ('organization', 'folder', 'project');
```

This Schema Defintion Generates the following directory structure and files.

```bash 
sqlcgen
├── baz
│   └── bar
│       └── foo
│           └── v1
│               └── message.proto
├── enum
│   └── v1
│       └── enum.proto
├── ets
│   └── api
│       ├── enum
│       │   └── v1
│       │       └── enum.proto
│       ├── iam
│       │   └── v1
│       │       └── message.proto
│       └── resourcemanager
│           └── v1
│               └── message.proto
├── foo
│   └── bar
│       └── baz
│           └── v1
│               └── message.proto
└── the
    └── resourcemanager
        └── v1
            └── message.proto
```

sqlcgen/foo/bar/baz/v1/message.proto:
```proto
syntax = "proto3";

package foo.bar.baz.v1;

import "google/protobuf/wrappers";
import "google/protobuf/timestamp";

message Users {
  bytes uuid = 0;

  google.protobuf.StringValue alias = 1;

  google.protobuf.Timestamp created_at = 2;

  google.protobuf.Timestamp deleted_at = 3;

  google.protobuf.StringValue description = 4;

  repeated string memberof = 5;

  string name = 6;

  google.protobuf.Timestamp updated_at = 7;
}
```

sqlcgen/baz/bar/foo/v1/message.proto
```proto
syntax = "proto3";

package baz.bar.foo.v1;

import "google/protobuf/wrappers";
import "google/protobuf/timestamp";

message Groups {
  bytes uuid = 0;

  google.protobuf.StringValue alias = 1;

  google.protobuf.Timestamp created_at = 2;

  google.protobuf.Timestamp deleted_at = 3;

  google.protobuf.StringValue description = 4;

  string name = 5;

  google.protobuf.Timestamp updated_at = 6;
}
```

sqlcgen/the/resourcemanager/v1/message.proto
```proto
syntax = "proto3";

package the.resourcemanager.v1;

import "enum/v1/enum.proto";
import "google/protobuf/wrappers";
import "google/protobuf/timestamp";

message Resource {
  google.protobuf.Timestamp created_at = 0;

  google.protobuf.Timestamp deleted_at = 1;

  google.protobuf.BytesValue parent_resource_uuid = 2;

  string resource_name = 3;

  enum.v1.ResourceType resource_type = 4;

  bytes resource_uuid = 5;

  google.protobuf.Timestamp updated_at = 6;
}
```

sqlcgen/enum/v1/enum.proto
```proto
syntax = "proto3";

package enum.v1;


enum ResourceType {
  RESOURCE_TYPE_UNSPECIFIED = 0;

  RESOURCE_TYPE_ORGANIZATION = 1;

  RESOURCE_TYPE_FOLDER = 2;

  RESOURCE_TYPE_PROJECT = 3;
}
```



#### Extending Messages, Enums, Request_Response and Services:
Messages defined in the user_defined directory that match the generate file path will be merged or appended.

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
```
