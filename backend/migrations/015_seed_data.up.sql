-- Seed Data: Demo users, posts, follows, likes, bookmarks
-- All users have password: password123
-- bcrypt hash: $2a$10$6yKkR29M2Nl9zUAdYKMk5.ZgCU.6sURy1aG2SVMYPx.7sUtaRc6QK

-- ============================================================
-- 1. Users (5명)
-- ============================================================
INSERT INTO users (id, email, password_hash, username, display_name, bio, profile_image_url, header_image_url)
VALUES
  ('a0000000-0000-0000-0000-000000000001', 'alice@example.com', '$2a$10$6yKkR29M2Nl9zUAdYKMk5.ZgCU.6sURy1aG2SVMYPx.7sUtaRc6QK',
   'alice', 'Alice Kim', 'Full-stack developer. Love coffee and code.', '', ''),
  ('a0000000-0000-0000-0000-000000000002', 'bob@example.com', '$2a$10$6yKkR29M2Nl9zUAdYKMk5.ZgCU.6sURy1aG2SVMYPx.7sUtaRc6QK',
   'bob', 'Bob Park', 'Designer & photographer. Based in Seoul.', '', ''),
  ('a0000000-0000-0000-0000-000000000003', 'charlie@example.com', '$2a$10$6yKkR29M2Nl9zUAdYKMk5.ZgCU.6sURy1aG2SVMYPx.7sUtaRc6QK',
   'charlie', 'Charlie Lee', 'Backend engineer at startup. Go enthusiast.', '', ''),
  ('a0000000-0000-0000-0000-000000000004', 'diana@example.com', '$2a$10$6yKkR29M2Nl9zUAdYKMk5.ZgCU.6sURy1aG2SVMYPx.7sUtaRc6QK',
   'diana', 'Diana Choi', 'Product manager. Building cool things.', '', ''),
  ('a0000000-0000-0000-0000-000000000005', 'eve@example.com', '$2a$10$6yKkR29M2Nl9zUAdYKMk5.ZgCU.6sURy1aG2SVMYPx.7sUtaRc6QK',
   'eve', 'Eve Jung', 'Data scientist. ML & AI.', '', '')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 2. Follows
-- ============================================================
INSERT INTO follows (follower_id, following_id) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002'), -- alice → bob
  ('a0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003'), -- alice → charlie
  ('a0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000004'), -- alice → diana
  ('a0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001'), -- bob → alice
  ('a0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000003'), -- bob → charlie
  ('a0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001'), -- charlie → alice
  ('a0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001'), -- diana → alice
  ('a0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000002'), -- diana → bob
  ('a0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000003'), -- diana → charlie
  ('a0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000005'), -- diana → eve
  ('a0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001')  -- eve → alice
ON CONFLICT DO NOTHING;

-- ============================================================
-- 3. Posts (public)
-- ============================================================
INSERT INTO posts (id, author_id, content, visibility, view_count, created_at, updated_at) VALUES
  -- alice 게시물
  ('b0000000-0000-0000-0000-000000000001',
   'a0000000-0000-0000-0000-000000000001',
   'Hello world! This is my first post on X Clone. Excited to be here! 🎉',
   'public', 42,
   NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),

  ('b0000000-0000-0000-0000-000000000002',
   'a0000000-0000-0000-0000-000000000001',
   E'## Today I learned\n\nYou can use **markdown** in posts now!\n\n- Bold text\n- `inline code`\n- [Links](https://example.com)\n\nPretty cool, right?',
   'public', 128,
   NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),

  -- bob 게시물
  ('b0000000-0000-0000-0000-000000000003',
   'a0000000-0000-0000-0000-000000000002',
   'Just finished a new design project. Minimalism is key.',
   'public', 35,
   NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),

  ('b0000000-0000-0000-0000-000000000004',
   'a0000000-0000-0000-0000-000000000002',
   E'Top 3 design tools for 2026:\n\n1. Figma\n2. Framer\n3. Spline\n\nWhat are yours?',
   'public', 89,
   NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),

  -- charlie 게시물
  ('b0000000-0000-0000-0000-000000000005',
   'a0000000-0000-0000-0000-000000000003',
   E'```go\nfunc main() {\n    fmt.Println("Hello, Go!")\n}\n```\n\nGo is beautiful in its simplicity.',
   'public', 210,
   NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),

  ('b0000000-0000-0000-0000-000000000006',
   'a0000000-0000-0000-0000-000000000003',
   'Hot take: error handling in Go is actually good. Fight me.',
   'public', 156,
   NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),

  -- diana 게시물
  ('b0000000-0000-0000-0000-000000000007',
   'a0000000-0000-0000-0000-000000000004',
   'Just shipped a new feature to production. The team did an amazing job!',
   'public', 67,
   NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),

  -- eve 게시물
  ('b0000000-0000-0000-0000-000000000008',
   'a0000000-0000-0000-0000-000000000005',
   E'Interesting paper on transformer architectures:\n\n> Attention is all you need, but *context* is what you make of it.\n\nThoughts?',
   'public', 93,
   NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 4. Posts (follower-only)
-- ============================================================
INSERT INTO posts (id, author_id, content, visibility, view_count, created_at, updated_at) VALUES
  ('b0000000-0000-0000-0000-000000000009',
   'a0000000-0000-0000-0000-000000000001',
   'This post is only visible to my followers. Testing visibility!',
   'follower', 12,
   NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),

  ('b0000000-0000-0000-0000-000000000010',
   'a0000000-0000-0000-0000-000000000003',
   'Follower-only post: Working on something secret... stay tuned.',
   'follower', 8,
   NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 5. Replies (parent_id)
-- ============================================================
INSERT INTO posts (id, author_id, content, visibility, parent_id, view_count, created_at, updated_at) VALUES
  -- bob replies to alice's first post
  ('c0000000-0000-0000-0000-000000000001',
   'a0000000-0000-0000-0000-000000000002',
   'Welcome to X Clone! Great to have you here.',
   'public', 'b0000000-0000-0000-0000-000000000001', 15,
   NOW() - INTERVAL '7 days' + INTERVAL '1 hour', NOW() - INTERVAL '7 days' + INTERVAL '1 hour'),

  -- charlie replies to alice's first post
  ('c0000000-0000-0000-0000-000000000002',
   'a0000000-0000-0000-0000-000000000003',
   'Hey Alice! Welcome aboard.',
   'public', 'b0000000-0000-0000-0000-000000000001', 10,
   NOW() - INTERVAL '7 days' + INTERVAL '2 hours', NOW() - INTERVAL '7 days' + INTERVAL '2 hours'),

  -- alice replies to charlie's Go post
  ('c0000000-0000-0000-0000-000000000003',
   'a0000000-0000-0000-0000-000000000001',
   'Go is my favorite language too! The concurrency model is amazing.',
   'public', 'b0000000-0000-0000-0000-000000000005', 22,
   NOW() - INTERVAL '4 days' + INTERVAL '3 hours', NOW() - INTERVAL '4 days' + INTERVAL '3 hours'),

  -- diana replies to charlie's hot take
  ('c0000000-0000-0000-0000-000000000004',
   'a0000000-0000-0000-0000-000000000004',
   'Agreed! Explicit error handling > hidden exceptions any day.',
   'public', 'b0000000-0000-0000-0000-000000000006', 18,
   NOW() - INTERVAL '2 days' + INTERVAL '5 hours', NOW() - INTERVAL '2 days' + INTERVAL '5 hours'),

  -- nested reply: charlie replies to alice's reply
  ('c0000000-0000-0000-0000-000000000005',
   'a0000000-0000-0000-0000-000000000003',
   'Right? Goroutines + channels = chef kiss',
   'public', 'c0000000-0000-0000-0000-000000000003', 9,
   NOW() - INTERVAL '4 days' + INTERVAL '4 hours', NOW() - INTERVAL '4 days' + INTERVAL '4 hours')
ON CONFLICT (id) DO NOTHING;

-- Update reply counts
UPDATE posts SET reply_count = 2 WHERE id = 'b0000000-0000-0000-0000-000000000001';
UPDATE posts SET reply_count = 1 WHERE id = 'b0000000-0000-0000-0000-000000000005';
UPDATE posts SET reply_count = 1 WHERE id = 'b0000000-0000-0000-0000-000000000006';
UPDATE posts SET reply_count = 1 WHERE id = 'c0000000-0000-0000-0000-000000000003';

-- ============================================================
-- 6. Likes
-- ============================================================
INSERT INTO likes (user_id, post_id) VALUES
  ('a0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001'), -- bob likes alice post 1
  ('a0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001'), -- charlie likes alice post 1
  ('a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'), -- diana likes alice post 1
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000005'), -- alice likes charlie go post
  ('a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000005'), -- diana likes charlie go post
  ('a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000005'), -- eve likes charlie go post
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000006'), -- alice likes charlie hot take
  ('a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000006'), -- diana likes charlie hot take
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000003'), -- alice likes bob design post
  ('a0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000004'), -- charlie likes bob tools post
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000008'), -- alice likes eve paper post
  ('a0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002'), -- bob likes alice markdown post
  ('a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000002')  -- eve likes alice markdown post
ON CONFLICT DO NOTHING;

-- Update like counts
UPDATE posts SET like_count = 3 WHERE id = 'b0000000-0000-0000-0000-000000000001';
UPDATE posts SET like_count = 2 WHERE id = 'b0000000-0000-0000-0000-000000000002';
UPDATE posts SET like_count = 1 WHERE id = 'b0000000-0000-0000-0000-000000000003';
UPDATE posts SET like_count = 1 WHERE id = 'b0000000-0000-0000-0000-000000000004';
UPDATE posts SET like_count = 3 WHERE id = 'b0000000-0000-0000-0000-000000000005';
UPDATE posts SET like_count = 2 WHERE id = 'b0000000-0000-0000-0000-000000000006';
UPDATE posts SET like_count = 1 WHERE id = 'b0000000-0000-0000-0000-000000000008';

-- ============================================================
-- 7. Bookmarks (alice)
-- ============================================================
INSERT INTO bookmarks (user_id, post_id) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000005'), -- alice bookmarks charlie go post
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000008'), -- alice bookmarks eve paper post
  ('a0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002')  -- bob bookmarks alice markdown post
ON CONFLICT DO NOTHING;
