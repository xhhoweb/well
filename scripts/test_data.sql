-- WellCMS Go 测试数据
-- 版块结构：
-- fid=1: 目录A (1篇帖子)
-- fid=2: 目录B (2篇帖子)
-- fid=3: 目录C (3篇帖子)
-- fid=4: 测试板块 (父版块)
-- fid=5: 测试子目录1 (4篇帖子)
-- fid=6: 测试子目录2 (5篇帖子)

-- ============================================
-- 1. 清空现有测试版块（可选执行）
-- ============================================
-- DELETE FROM thread WHERE fid IN (1,2,3,4,5,6);
-- DELETE FROM forum WHERE fid IN (1,2,3,4,5,6);

-- ============================================
-- 2. 创建版块
-- ============================================

-- 目录A (fid=1)
INSERT IGNORE INTO forum (fid, name, parent, path, depth, `order`, threads, today, posts, status)
VALUES (1, '目录A', 0, '0', 0, 1, 0, 0, 0, 0);

-- 目录B (fid=2)
INSERT IGNORE INTO forum (fid, name, parent, path, depth, `order`, threads, today, posts, status)
VALUES (2, '目录B', 0, '0', 0, 2, 0, 0, 0, 0);

-- 目录C (fid=3)
INSERT IGNORE INTO forum (fid, name, parent, path, depth, `order`, threads, today, posts, status)
VALUES (3, '目录C', 0, '0', 0, 3, 0, 0, 0, 0);

-- 测试板块 (fid=4)
INSERT IGNORE INTO forum (fid, name, parent, path, depth, `order`, threads, today, posts, status)
VALUES (4, '测试板块', 0, '0', 0, 4, 0, 0, 0, 0);

-- 测试子目录1 (fid=5)
INSERT IGNORE INTO forum (fid, name, parent, path, depth, `order`, threads, today, posts, status)
VALUES (5, '测试子目录1', 4, '0,4', 1, 1, 0, 0, 0, 0);

-- 测试子目录2 (fid=6)
INSERT IGNORE INTO forum (fid, name, parent, path, depth, `order`, threads, today, posts, status)
VALUES (6, '测试子目录2', 4, '0,4', 1, 2, 0, 0, 0, 0);

-- ============================================
-- 3. 创建帖子
-- ============================================

-- 目录A: 1篇帖子
INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (101, 1, 1, '目录A的第一篇帖子', 10, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (101, '这是目录A的第一篇帖子的内容。');

-- 目录B: 2篇帖子
INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (201, 2, 1, '目录B的第一篇帖子', 15, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (201, '这是目录B的第一篇帖子的内容。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (202, 2, 1, '目录B的第二篇帖子', 20, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (202, '这是目录B的第二篇帖子的内容，回复数更多。');

-- 目录C: 3篇帖子
INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (301, 3, 1, '目录C的第一篇帖子', 25, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (301, '这是目录C的第一篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (302, 3, 1, '目录C的第二篇帖子', 30, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (302, '这是目录C的第二篇帖子，有2个回复。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (303, 3, 1, '目录C的第三篇帖子', 35, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (303, '这是目录C的第三篇帖子，有3个回复。');

-- 测试子目录1: 4篇帖子
INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (501, 5, 1, '测试子目录1的第一篇帖子', 40, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (501, '这是测试子目录1的第一篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (502, 5, 1, '测试子目录1的第二篇帖子', 45, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (502, '这是测试子目录1的第二篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (503, 5, 1, '测试子目录1的第三篇帖子', 50, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (503, '这是测试子目录1的第三篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (504, 5, 1, '测试子目录1的第四篇帖子', 55, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (504, '这是测试子目录1的第四篇帖子。');

-- 测试子目录2: 5篇帖子
INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (601, 6, 1, '测试子目录2的第一篇帖子', 60, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (601, '这是测试子目录2的第一篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (602, 6, 1, '测试子目录2的第二篇帖子', 65, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (602, '这是测试子目录2的第二篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (603, 6, 1, '测试子目录2的第三篇帖子', 70, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (603, '这是测试子目录2的第三篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (604, 6, 1, '测试子目录2的第四篇帖子', 75, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (604, '这是测试子目录2的第四篇帖子。');

INSERT IGNORE INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status)
VALUES (605, 6, 1, '测试子目录2的第五篇帖子', 80, 4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);
INSERT IGNORE INTO thread_data (tid, message) VALUES (605, '这是测试子目录2的第五篇帖子，回复数最多。');

-- ============================================
-- 4. 创建标签
-- ============================================

INSERT IGNORE INTO tag (tag_id, name, slug, threads, view, status)
VALUES (1, '热门', 'hot', 10, 0, 0);

INSERT IGNORE INTO tag (tag_id, name, slug, threads, view, status)
VALUES (2, '精华', 'essence', 5, 0, 0);

INSERT IGNORE INTO tag (tag_id, name, slug, threads, view, status)
VALUES (3, '测试', 'test', 15, 0, 0);

-- ============================================
-- 5. 创建标签关联
-- ============================================

INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (101, 1);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (201, 1);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (201, 3);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (202, 2);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (202, 3);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (301, 1);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (302, 2);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (303, 3);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (501, 1);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (502, 2);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (503, 3);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (601, 1);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (602, 3);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (603, 2);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (604, 3);
INSERT IGNORE INTO thread_tag (tid, tag_id) VALUES (605, 1);
