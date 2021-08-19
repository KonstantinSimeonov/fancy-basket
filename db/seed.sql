insert into categories (created_at, updated_at, name) values
    (now(), now(), 'razni'),
    (now(), now(), 'markovi'),
    (now(), now(), 'obuvki'),
    (now(), now(), 'posuda');

do $$
declare
    ids uuid[] := array(select id from categories);
    x uuid;
    name text;
begin
    foreach x in array ids loop
        for i in 1..15 loop
            insert into products (created_at, updated_at, name, price, category_id, stock)
            values (now(), now(), 'neshto si ' || i, i * 2.2, x, 10);
        end loop;
    end loop;
end; $$
