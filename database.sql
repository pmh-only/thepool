CREATE TABLE collections (
  collection_id VARCHAR(10) PRIMARY KEY,
  file_original_name VARCHAR(255) NOT NULL,
  file_mime_type VARCHAR(255) NOT NULL,
  chunk_ids VARCHAR(10000) NOT NULL
) ENGINE=InnoDB;

CREATE TABLE chunks (
  chunk_id VARCHAR(10) PRIMARY KEY,
  chunk_size BIGINT NOT NULL,
  chunk_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  chunk_order INT UNSIGNED NOT NULL,
  KEY idx_chunks_order (chunk_created_at, chunk_order, chunk_id)
) ENGINE=InnoDB;

CREATE TABLE totals (
  id TINYINT PRIMARY KEY,
  total BIGINT NOT NULL DEFAULT 0
) ENGINE=InnoDB;

INSERT INTO totals (id,total) VALUES (1,0)
ON DUPLICATE KEY UPDATE total = VALUES(total);

DELIMITER $$

CREATE TRIGGER chunks_ai AFTER INSERT ON chunks
FOR EACH ROW
BEGIN
  UPDATE totals SET total = total + NEW.chunk_size WHERE id = 1;
END$$

CREATE TRIGGER chunks_ad AFTER DELETE ON chunks
FOR EACH ROW
BEGIN
  UPDATE totals SET total = total - OLD.chunk_size WHERE id = 1;
END$$

CREATE TRIGGER chunks_au AFTER UPDATE ON chunks
FOR EACH ROW
BEGIN
  UPDATE totals
    SET total = total + (NEW.chunk_size - OLD.chunk_size)
  WHERE id = 1;
END$$

DROP PROCEDURE IF EXISTS purge_chunks_to_limit$$
CREATE PROCEDURE purge_chunks_to_limit(IN p_limit BIGINT UNSIGNED)
proc:BEGIN
  DECLARE v_total  BIGINT UNSIGNED;
  DECLARE v_excess BIGINT UNSIGNED;

  START TRANSACTION;

  SELECT total INTO v_total FROM totals WHERE id = 1 FOR UPDATE;
  SET v_excess = IF(v_total > p_limit, v_total - p_limit, 0);

  IF v_excess = 0 THEN
    COMMIT;
    LEAVE proc;
  END IF;

  DROP TEMPORARY TABLE IF EXISTS tmp_to_delete;
  CREATE TEMPORARY TABLE tmp_to_delete (
    chunk_id VARCHAR(10) PRIMARY KEY
  ) ENGINE=Memory;

  INSERT INTO tmp_to_delete (chunk_id)
  SELECT chunk_id
  FROM (
    SELECT
      chunk_id,
      SUM(chunk_size) OVER (ORDER BY chunk_created_at, chunk_order, chunk_id) AS run_sum,
      COALESCE(
        SUM(chunk_size) OVER (
          ORDER BY chunk_created_at, chunk_order, chunk_id
          ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
        ), 0
      ) AS prev_sum
    FROM chunks
  ) s
  WHERE run_sum <= v_excess
     OR (prev_sum < v_excess AND run_sum >= v_excess);

  DROP TEMPORARY TABLE IF EXISTS tmp_deleted;
  CREATE TEMPORARY TABLE tmp_deleted AS
    SELECT c.chunk_id, c.chunk_created_at, c.chunk_order
    FROM chunks c
    JOIN tmp_to_delete d USING (chunk_id);

  DELETE FROM chunks
  WHERE chunk_id IN (SELECT chunk_id FROM tmp_to_delete);

  COMMIT;

  SELECT chunk_id
  FROM tmp_deleted
  ORDER BY chunk_created_at, chunk_order, chunk_id;
END$$

DELIMITER ;
