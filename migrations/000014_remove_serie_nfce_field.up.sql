-- Remove serie_nfce field from companies table
-- This field is no longer used since the system only supports serie '1'

ALTER TABLE companies DROP COLUMN IF EXISTS serie_nfce;
