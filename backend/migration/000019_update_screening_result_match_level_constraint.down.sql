-- Rollback: Remove 'no_match' from match_level constraint
ALTER TABLE "screening_results"
DROP CONSTRAINT IF EXISTS "chk_screening_results_match_level";

ALTER TABLE "screening_results"
ADD CONSTRAINT "chk_screening_results_match_level" CHECK ("match_level" IN ('excellent', 'good', 'fair', 'poor'));

-- Restore original comment
COMMENT ON COLUMN "screening_results"."match_level" IS '匹配等级: excellent/good/fair/poor';