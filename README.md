# GeoDB

[![Github licence](https://img.shields.io/github/license/tenrok/geodb)](https://raw.githubusercontent.com/tenrok/geodb/main/LICENSE)
[![Example](https://img.shields.io/badge/example-blue)](https://github.com/tenrok/GeoDB/tree/example)

База данных для определения геолокации по IPv4 в формате SQLite3

[Скачать GeoDB.sqlite](https://github.com/tenrok/GeoDB/raw/main/GeoDB.sqlite)

#### Пример запроса к БД

```sql
with ip2long(ip, long) as (
  select '62.212.64.19', 0
  union all
  select
    ip,
    cast(replace(ip, ltrim(ip, '1234567890'), '') * 16777216
      + rtrim(rtrim(rtrim(ltrim(ltrim(ip, '1234567890'), '.'), '1234567890'), '.'), '1234567890') * 65536
      + ltrim(ltrim(rtrim(ltrim(ltrim(ip, '1234567890'), '.'), '1234567890'), '1234567890'), '.') * 256
      + replace(ip, rtrim(ip, '1234567890'), '') as integer
    ) long
  from ip2long where long = 0
)
select
  n.ip,
  k.iso "continent_code",
  k.name_en "continent_name_en",
  k.name_ru "continent_name_ru",
  s.iso "country_iso",
  s.name_en "country_name_en",
  s.name_ru "country_name_ru",
  r.iso "region_code",
  r.name_en "region_name_en",
  r.name_ru "region_name_ru",
  g.name_en "city_name_en",
  g.name_ru "city_name_ru",
  g.lat "city_lat",
  g.lon "city_lon",
  r.timezone "region_timezone"
from networks n
  left join countries s on s.id = n.country_id
  left join continents k on k.id = s.continent_id
  left join regions r on r.id = n.region_id
  left join cities g on g.id = n.city_id
where
  (select long from ip2long i where i.long != 0) >= n.ip
order by ip desc
limit 1;
```
