-- name: CreateOTP :one
INSERT INTO otp_codes (phone, code, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '5 minutes')
RETURNING *;

-- name: GetValidOTP :one
SELECT * FROM otp_codes
WHERE phone = $1
  AND code = $2
  AND verified_at IS NULL
  AND attempts < max_attempts
  AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementOTPAttempts :exec
UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1;

-- name: MarkOTPVerified :exec
UPDATE otp_codes SET verified_at = NOW() WHERE id = $1;

-- name: CleanExpiredOTPs :exec
DELETE FROM otp_codes WHERE expires_at < NOW() - INTERVAL '1 hour';

-- name: CountRecentOTPs :one
SELECT COUNT(*) FROM otp_codes
WHERE phone = $1 AND created_at > NOW() - INTERVAL '1 hour';

-- name: GetLatestOTPByPhone :one
SELECT * FROM otp_codes
WHERE phone = $1
  AND verified_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;
