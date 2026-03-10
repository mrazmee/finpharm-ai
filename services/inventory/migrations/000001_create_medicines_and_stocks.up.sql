CREATE TABLE IF NOT EXISTS medicines (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stocks (
    medicine_id VARCHAR(50) PRIMARY KEY REFERENCES medicines(id) ON DELETE CASCADE,
    available_qty INTEGER NOT NULL CHECK (available_qty >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_medicines_type ON medicines(type);

INSERT INTO medicines (id, name, type)
VALUES
    ('PARA500', 'Paracetamol 500mg', 'OTC'),
    ('AMOX500', 'Amoxicillin 500mg', 'ANTIBIOTIC'),
    ('OBATKERAS-X', 'Obat Keras X', 'CONTROLLED')
ON CONFLICT (id) DO NOTHING;

INSERT INTO stocks (medicine_id, available_qty)
VALUES
    ('PARA500', 80),
    ('AMOX500', 120),
    ('OBATKERAS-X', 5)
ON CONFLICT (medicine_id) DO NOTHING;