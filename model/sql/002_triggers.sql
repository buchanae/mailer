create trigger if not exists increment_next_message_id after insert on message
for each row
begin
  update mailbox set next_message_id = next_message_id + 1 where id = new.mailbox_id;
end;

create trigger if not exists set_seen_index after insert on flag
for each row
when new.value = '\seen'
begin
  update message set seen = 1 where row_id = new.message_row_id;
end;

create trigger if not exists set_recent_index after insert on flag
for each row
when new.value = '\recent'
begin
  update message set recent = 1 where row_id = new.message_row_id;
end;

create trigger if not exists set_answered_index after insert on flag
for each row
when new.value = '\answered'
begin
  update message set answered = 1 where row_id = new.message_row_id;
end;

create trigger if not exists set_flagged_index after insert on flag
for each row
when new.value = '\flagged'
begin
  update message set flagged = 1 where row_id = new.message_row_id;
end;

create trigger if not exists set_draft_index after insert on flag
for each row
when new.value = '\draft'
begin
  update message set draft = 1 where row_id = new.message_row_id;
end;

create trigger if not exists set_deleted_index after insert on flag
for each row
when new.value = '\deleted'
begin
  update message set deleted = 1 where row_id = new.message_row_id;
end;

create trigger if not exists unset_seen_index after delete on flag
for each row
when old.value = '\seen'
begin
  update message set seen = 0 where row_id = old.message_row_id;
end;

create trigger if not exists unset_recent_index after delete on flag
for each row
when old.value = '\recent'
begin
  update message set recent = 0 where row_id = old.message_row_id;
end;

create trigger if not exists unset_answered_index after delete on flag
for each row
when old.value = '\answered'
begin
  update message set answered = 0 where row_id = old.message_row_id;
end;

create trigger if not exists unset_flagged_index after delete on flag
for each row
when old.value = '\flagged'
begin
  update message set flagged = 0 where row_id = old.message_row_id;
end;

create trigger if not exists unset_draft_index after delete on flag
for each row
when old.value = '\draft'
begin
  update message set draft = 0 where row_id = old.message_row_id;
end;

create trigger if not exists unset_deleted_index after delete on flag
for each row
when old.value = '\deleted'
begin
  update message set deleted = 0 where row_id = old.message_row_id;
end;
