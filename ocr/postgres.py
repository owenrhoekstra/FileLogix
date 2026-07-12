"""
Shared Postgres connection pool for Python services.

Uses a Unix socket (matching the Go backend — Postgres runs on the host,
not containerized), so DB_HOST is intentionally not used. `host` is set
to the socket directory, not a hostname.

Usage:
    from postgres import get_connection

    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("SELECT ...")
"""

import os
from contextlib import contextmanager

from psycopg2 import pool

DB_USER = os.environ["DB_USER"]
DB_PASSWORD = os.environ["DB_PASSWORD"]
DB_NAME = os.environ["DB_NAME"]
DB_SOCKET_DIR = os.environ.get("DB_SOCKET_DIR", "/var/run/postgresql")

_pool = pool.ThreadedConnectionPool(
    minconn=1,
    maxconn=10,
    host=DB_SOCKET_DIR,
    dbname=DB_NAME,
    user=DB_USER,
    password=DB_PASSWORD,
)


@contextmanager
def get_connection():
    """Borrow a connection from the pool, always returning it when done."""
    conn = _pool.getconn()
    try:
        yield conn
    finally:
        _pool.putconn(conn)


def close_all():
    """Call on service shutdown to close every pooled connection cleanly."""
    _pool.closeall()