create table if not exists message (
  id integer not null primary key autoincrement,
  mailbox_id integer not null references mailbox(id) on delete cascade on update cascade,

  size integer not null default 0,

  seen integer not null default 0,
  recent integer not null default 0,
  answered integer not null default 0,
  flagged integer not null default 0,
  draft integer not null default 0,
  deleted integer not null default 0,

  created datetime not null,
  path text not null default ''
);

create index if not exists message_seen_index on message (seen);
create index if not exists message_answered_index on message (answered);
create index if not exists message_recent_index on message (recent);
create index if not exists message_flagged_index on message (flagged);
create index if not exists message_draft_index on message (draft);
create index if not exists message_deleted_index on message (deleted);

create table if not exists header (
  message_id integer not null references message(id) on delete cascade on update cascade,

  key text not null,
  value text not null
);

create index if not exists header_key_index on header (key);

create table if not exists flag (
  message_id integer not null references message(id) on delete cascade on update cascade,

  value text not null,

  primary key (message_id, value collate nocase)
);

create table if not exists mailbox (
  id integer not null primary key autoincrement,

  name text not null,

  unique (name collate nocase)
);
