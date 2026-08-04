-- Backfill: strip frozen presigned URLs out of stored berita content.
-- normalizeImageRefs (write path) previously failed to convert path-style
-- signed URLs (https://s3.<region>.backblazeb2.com/<bucket>/<key>) back to
-- bare object keys, so expiring URLs were persisted verbatim. This rewrites
-- any path-style B2 URL back to its bare key.
UPDATE berita
SET content = regexp_replace(
    content,
    'https://[^/]+\.backblazeb2\.com/jhic-berita-images/(berita/[0-9]+/content/[^)?\s]+)\?[^)]*',
    '\1',
    'g'
)
WHERE content ~ 'https://[^/]+\.backblazeb2\.com/jhic-berita-images/berita/';
