ALTER TABLE `session` RENAME COLUMN `title` to `entry_title`;
ALTER TABLE `session` ADD COLUMN `exit_title` String;
