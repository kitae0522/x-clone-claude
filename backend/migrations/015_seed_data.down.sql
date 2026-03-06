-- Remove seed data (reverse order of insertion)
DELETE FROM bookmarks WHERE user_id IN (
  'a0000000-0000-0000-0000-000000000001',
  'a0000000-0000-0000-0000-000000000002'
) AND post_id LIKE 'b0000000%';

DELETE FROM likes WHERE user_id LIKE 'a0000000%' AND post_id LIKE 'b0000000%';

DELETE FROM posts WHERE id LIKE 'c0000000%';
DELETE FROM posts WHERE id LIKE 'b0000000%';

DELETE FROM follows WHERE follower_id LIKE 'a0000000%';

DELETE FROM users WHERE id LIKE 'a0000000%';
