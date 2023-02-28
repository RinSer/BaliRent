alter table public.properties
    alter column price integer int8 using price::integer;