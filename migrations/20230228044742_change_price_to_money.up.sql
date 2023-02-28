alter table public.properties
    alter column price type numeric(19,2) using price::numeric(19,2);