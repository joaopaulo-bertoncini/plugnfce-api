-- Drop NFC-e sequences table and function
DROP FUNCTION IF EXISTS get_next_nfce_number(UUID, VARCHAR(3));
DROP TABLE IF EXISTS nfce_sequences;
