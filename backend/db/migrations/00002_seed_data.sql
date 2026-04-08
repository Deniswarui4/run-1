-- +goose Up
-- +goose StatementBegin

-- Default platform settings (only insert if no record exists)
INSERT INTO platform_settings (
    platform_fee_percentage,
    withdrawal_fee_percentage,
    min_withdrawal_amount,
    currency
)
SELECT 5.0, 2.5, 1000, 'KES'
WHERE NOT EXISTS (SELECT 1 FROM platform_settings)
ON CONFLICT DO NOTHING;

-- Default categories
INSERT INTO categories (name, description, color, icon, is_active) VALUES
    ('Music',          'Concerts, festivals, and live music events',          '#8B5CF6', 'music',          TRUE),
    ('Sports',         'Sporting events, tournaments, and fitness activities', '#EF4444', 'sports',         TRUE),
    ('Arts & Culture', 'Art exhibitions, theatre, and cultural experiences',  '#F59E0B', 'arts',           TRUE),
    ('Technology',     'Tech conferences, hackathons, and workshops',         '#3B82F6', 'technology',     TRUE),
    ('Business',       'Networking events, seminars, and trade shows',        '#10B981', 'business',       TRUE),
    ('Food & Drink',   'Food festivals, tastings, and culinary experiences',  '#F97316', 'food',           TRUE),
    ('Education',      'Workshops, training sessions, and lectures',          '#6366F1', 'education',      TRUE),
    ('Health',         'Wellness events, fitness classes, and health expos',  '#14B8A6', 'health',         TRUE),
    ('Comedy',         'Stand-up shows, improv nights, and comedy festivals', '#FBBF24', 'comedy',         TRUE),
    ('Charity',        'Fundraisers, community events, and charity galas',    '#EC4899', 'charity',        TRUE)
ON CONFLICT (name) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove seeded categories
DELETE FROM categories
WHERE name IN (
    'Music', 'Sports', 'Arts & Culture', 'Technology', 'Business',
    'Food & Drink', 'Education', 'Health', 'Comedy', 'Charity'
);

-- Remove seeded platform settings (only if updated_by is NULL, i.e. never modified by an admin)
DELETE FROM platform_settings WHERE updated_by IS NULL;

-- +goose StatementEnd
