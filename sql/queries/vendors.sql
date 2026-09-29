-- name: CreateVendor :one
INSERT INTO vendors (user_id, name, city)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVendorByUserID :one
SELECT * FROM vendors WHERE user_id = $1;

-- name: GetVendorByID :one
SELECT * FROM vendors WHERE id = $1;

-- name: UpdateVendor :one
UPDATE vendors
SET name = $2,
    city = $3,
    pic_name = $4,
    whatsapp = $5,
    address = $6,
    province = $7,
    province_code = $8,
    city_code = $9,
    district = $10,
    district_code = $11,
    photo_url = $12,
    bio = $13,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;
