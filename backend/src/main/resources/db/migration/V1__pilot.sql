CREATE EXTENSION IF NOT EXISTS postgis;
CREATE TABLE account (
 id uuid PRIMARY KEY, email text UNIQUE NOT NULL, password_hash text NOT NULL,
 display_name text NOT NULL, location geography(Point,4326) NOT NULL
);
CREATE TABLE session_token (
 token_hash text PRIMARY KEY, account_id uuid NOT NULL REFERENCES account(id),
 expires_at timestamptz NOT NULL
);
CREATE TABLE book (
 id uuid PRIMARY KEY, owner_id uuid NOT NULL REFERENCES account(id),
 title text NOT NULL, author text NOT NULL, description text,
 rental_price_paise integer NOT NULL CHECK(rental_price_paise>=0),
 archived boolean NOT NULL DEFAULT false
);
CREATE INDEX account_location_idx ON account USING gist(location);
CREATE TABLE borrow_request (
 id uuid PRIMARY KEY, book_id uuid NOT NULL REFERENCES book(id),
 borrower_id uuid NOT NULL REFERENCES account(id), note text,
 status text NOT NULL CHECK(status IN ('PENDING','ACCEPTED','DECLINED','CANCELLED','RETURNED')),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX one_active_loan ON borrow_request(book_id) WHERE status='ACCEPTED';
CREATE UNIQUE INDEX one_pending_per_borrower ON borrow_request(book_id,borrower_id) WHERE status='PENDING';
CREATE TABLE message (
 id uuid PRIMARY KEY, request_id uuid NOT NULL REFERENCES borrow_request(id),
 sender_id uuid NOT NULL REFERENCES account(id), body text NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX message_request_idx ON message(request_id,created_at,id);
CREATE TABLE user_block (
 blocker_id uuid REFERENCES account(id), blocked_id uuid REFERENCES account(id),
 PRIMARY KEY(blocker_id,blocked_id), CHECK(blocker_id<>blocked_id)
);
CREATE TABLE report (
 id uuid PRIMARY KEY, request_id uuid NOT NULL REFERENCES borrow_request(id),
 reporter_id uuid NOT NULL REFERENCES account(id), reported_id uuid NOT NULL REFERENCES account(id),
 reason text NOT NULL, details text, created_at timestamptz NOT NULL DEFAULT now(), reviewed boolean NOT NULL DEFAULT false
);
