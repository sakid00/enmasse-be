-- name: CreateArtist :one
INSERT INTO artists (
    user_id,
    first_name,
    last_name,
    artist_name,
    photo_url,
    address,
    website_url,
    instagram_handle,
    twitter_handle,
    whatsapp,
    imported_artist_id,
    imported_portal_user_id,
    bio,
    nationality
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: GetArtistByUserID :one
SELECT * FROM artists WHERE user_id = $1;

-- name: GetArtistByID :one
SELECT * FROM artists WHERE id = $1;

-- name: UpdateArtistProfile :one
UPDATE artists
SET
    first_name = $2,
    last_name = $3,
    artist_name = $4,
    photo_url = $5,
    address = $6,
    website_url = $7,
    instagram_handle = $8,
    twitter_handle = $9,
    whatsapp = $10,
    bio = $11,
    province = $12,
    province_code = $13,
    city = $14,
    city_code = $15,
    district = $16,
    district_code = $17,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;
