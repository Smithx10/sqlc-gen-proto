# sqlc-gen-proto Generates .proto files


### Current Comment Annotation Options:
#### -- generate:

#### -- package:

#### -- replace:

#### -- outdir:

#### -- filename:



```sql
-- generate:
CREATE TABLE "public"."users" (
  "uuid" uuid NOT NULL,
  "name" character varying NOT NULL,
  "alias" character varying NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NOT NULL,
  PRIMARY KEY ("uuid")
);


-- generate:
-- package: baz.baz.bar.v1
-- replace: description google/type/expr.proto google.type.Expr
CREATE TABLE "public"."GROUPS" (
  "uuid" uuid NOT NULL, 
  "NaMe" character varying NOT NULL,
  "AlIas" character varying NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NOT NULL,
  PRIMARY KEY ("uuid")
);

-- generate:
-- package: foo.bar.baz.v1
CREATE TABLE "public"."group_members" (
  "group_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("group_id", "user_id"),
  CONSTRAINT "group_members_group_id" FOREIGN KEY ("group_id") REFERENCES "public"."groups" ("uuid") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "group_members_user_id" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("uuid") ON UPDATE NO ACTION ON DELETE CASCADE
);

COMMENT ON TABLE "public"."group_members" IS 'Test';
```

This Schema Defintion Generates the following directory structure and files.

```bash 
sqlcgen
├── baz
│   └── baz
│       └── bar
│           └── v1
│               └── message.proto
├── foo
│   └── bar
│       └── baz
│           └── v1
│               └── message.proto
└── sqlcgen
    └── message.proto
```

sqlcgen/sqlcgen/message.proto:
```proto
syntax = "proto3";

package sqlcgen;

import "google/protobuf/wrappers";
import "google/protobuf/timestamp";

message Users  {

  bytes uuid = 0;

  string name = 1;

  google.protobuf.StringValue alias = 2;

  google.protobuf.StringValue description = 3;

  google.protobuf.Timestamp created_at = 4;

  google.protobuf.Timestamp updated_at = 5;

  google.protobuf.Timestamp deleted_at = 6;

}
```

sqlcgen/foo/bar/baz/v1/message.proto:
```proto
syntax = "proto3";

package foo.bar.baz.v1;


message GroupMembers  {

  bytes group_id = 0;

  bytes user_id = 1;

}
```

sqlcgen/baz/baz/bar/v1/message.proto:
```proto
syntax = "proto3";

package baz.baz.bar.v1;

import "google/protobuf/wrappers";
import "google/protobuf/timestamp";
import "google/type/expr.proto";

message Groups  {

  bytes uuid = 0;

  string name = 1;

  google.protobuf.StringValue alias = 2;

  google.type.Expr description = 3;

  google.protobuf.Timestamp created_at = 4;

  google.protobuf.Timestamp updated_at = 5;

  google.protobuf.Timestamp deleted_at = 6;

}
```



