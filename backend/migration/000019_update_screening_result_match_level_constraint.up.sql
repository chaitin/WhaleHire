-- Update match_level constraint to include 'no_match' value
ALTER TABLE "screening_results"
DROP CONSTRAINT IF EXISTS "chk_screening_results_match_level";

ALTER TABLE "screening_results"
ADD CONSTRAINT "chk_screening_results_match_level" CHECK ("match_level" IN ('excellent', 'good', 'fair', 'poor', 'no_match'));

-- Update comment to reflect the new constraint
COMMENT ON COLUMN "screening_results"."match_level" IS '匹配等级: excellent/good/fair/poor/no_match';