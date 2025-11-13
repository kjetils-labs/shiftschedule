## **Database Schema Overview**

This schema is designed for a **shift scheduling system**, where personnel can be assigned to specific schedules that belong to defined schedule types.  
It supports managing who can participate in which kinds of schedules, as well as tracking assigned shifts, substitutes, and acceptance status.

---

### **Table: `personnel`**

**Purpose:**  
Stores information about all personnel who can be scheduled for shifts.

**Columns:**

| Column | Type                 | Description                        |
| ------ | -------------------- | ---------------------------------- |
| `id`   | `serial primary key` | Unique identifier for each person. |
| `name` | `text not null`      | The person's full name.            |

**Usage Notes:**

- All users or employees that participate in scheduling must exist here.
- Each person can be linked to one or more schedule types via `schedule_type_personnel`.

---

### **Table: `schedule_type`**

**Purpose:**  
Defines the types of schedules available (e.g., "Primary TRD", "Backup Shift", "Night Duty").  
Acts as a _template or category_ for actual schedules.

**Columns:**

| Column        | Type                   | Description                                                |
| ------------- | ---------------------- | ---------------------------------------------------------- |
| `id`          | `serial primary key`   | Unique identifier for the schedule type.                   |
| `name`        | `text not null unique` | The unique name of the schedule type.                      |
| `description` | `text`                 | Optional description of what the schedule type represents. |

**Usage Notes:**

- Each `shiftschedule` references one `schedule_type`.
- You can control personnel eligibility per schedule type.

---

### **Table: `schedule_type_personnel`**

**Purpose:**  
Defines which personnel are allowed to participate in which schedule types.

**Columns:**

| Column             | Type                               | Description                                     |
| ------------------ | ---------------------------------- | ----------------------------------------------- |
| `schedule_type_id` | `int references schedule_type(id)` | The schedule type the personnel can work under. |
| `personnel_id`     | `int references personnel(id)`     | The person eligible for that schedule type.     |

**Constraints:**

- `PRIMARY KEY (schedule_type_id, personnel_id)` ensures unique pairs.
- `ON DELETE CASCADE` ensures relationships are cleaned up automatically.

**Usage Notes:**

- Use this table to enforce business rules on eligibility.
- Before assigning a person to a shift, ensure a row exists in this table for that type.

---

### **Table: `shiftschedule`**

**Purpose:**  
Represents an individual shift schedule — a concrete instance of a schedule type, tied to a specific week.

**Columns:**

| Column             | Type                               | Description                                                            |
| ------------------ | ---------------------------------- | ---------------------------------------------------------------------- |
| `id`               | `serial primary key`               | Unique ID for each schedule instance.                                  |
| `name`             | `text not null`                    | Name or label for the schedule instance (e.g., “Primary TRD Week 12”). |
| `weeknumber`       | `integer not null`                 | ISO week number for the schedule.                                      |
| `assignee`         | `int references personnel(id)`     | Primary person assigned to this shift.                                 |
| `substitute`       | `int references personnel(id)`     | Optional substitute for the shift.                                     |
| `comment`          | `text`                             | Optional free-text comment.                                            |
| `accepted`         | `boolean`                          | Indicates if the assignee has accepted the schedule.                   |
| `schedule_type_id` | `int references schedule_type(id)` | Links this schedule to its type.                                       |

**Usage Notes:**

- This table stores the actual scheduling data week-by-week.
- Each schedule belongs to one `schedule_type`.
- Personnel assignments should respect eligibility rules in `schedule_type_personnel`.

---

## **Schema Summary**

| Table                             | Role                                                              | Relationship Highlights                                 |
| --------------------------------- | ----------------------------------------------------------------- | ------------------------------------------------------- |
| `personnel`                       | Stores people who can work shifts.                                | Linked to schedule types via `schedule_type_personnel`. |
| `schedule_type`                   | Defines possible categories of schedules.                         | Linked to personnel and shift schedules.                |
| `schedule_type_personnel`         | Maps who can participate in which schedule types.                 | Many-to-many between `personnel` and `schedule_type`.   |
| `shiftschedule`                   | Concrete schedule instances (e.g., week-based shifts).            | Linked to `schedule_type` and assigned personnel.       |
| `schedule_personnel` _(optional)_ | Links personnel to specific schedules beyond assignee/substitute. | Many-to-many between `personnel` and `shiftschedule`.   |
