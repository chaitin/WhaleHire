CREATE TABLE "settings" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "enable_sso" boolean NOT NULL DEFAULT false,
    "force_two_factor_auth" boolean NOT NULL DEFAULT false,
    "disable_password_login" boolean NOT NULL DEFAULT false,
    "enable_auto_login" boolean NOT NULL DEFAULT false,
    "dingtalk_oauth" jsonb NULL,
    "custom_oauth" jsonb NULL,
    "base_url" character varying NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    PRIMARY KEY ("id")
);