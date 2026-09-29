# Enmasse

Enmasse is the identity and directory service. It owns login, password hashes, artist profiles, vendor core (name, city, PIC, street address, WhatsApp, province, district, bio), and whether a profile is complete enough for logged-in explore (artists) or staff review (vendors). City is collected on onboarding wilayah, not at register.

Massmaker stores only `enmasse_artist_id` / `enmasse_vendor_id`. It never copies emails or password hashes. The two services run on different VPS machines and talk over HTTPS with a short-lived service JWT.

`profile_complete` is computed. Artists need street address, province, city, district, bio, at least one social, and a +62 WhatsApp. Profile photo is optional. Vendors need name, province, city, district, PIC, street address, bio, and a +62 WhatsApp. `address` is the street line only. It is not a staff-flipped column.

Claim always returns 200. Seeded users have `password_hash` null until they set a password from a hashed, single-use token.
