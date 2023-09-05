-- window
select e.id, e.name, CASE WHEN s.bonus IS NOT NULL THEN (e.gross_salary+s.bonus)*0.87 ELSE 0 END AS last_salary from employee e
left join 
(select employee_id, bonus
row_number() over (partition by employee_id order by date desc) as rn
from new_schema.salary) s
on s.employee_id = e.id and s.rn = 1


-- subquery
with last_payments as (
	select employee_id, max(date) as last_date from new_schema.salary group by employee_id
    order by employee_id
)
select e.id, CASE WHEN s.bonus IS NOT NULL THEN (e.gross_salary+s.bonus)*0.87 ELSE 0 END AS last_salary
FROM new_schema.employee e
left join new_schema.salary s
on e.id = s.employee_id 
and s.date = (select last_date from last_payments where last_payments.employee_id = e.id);