-- These are the tables that make up our database

create table if not exists user (
    user_id   	 integer not null primary key autoincrement,
    username	 text    not null unique,
    passhash	 text    not null,
    email	 text    not null unique
);

create table if not exists submission_status (
    submission_status_id integer not null primary key autoincrement,
    status	 	 text    not null unique
);

create table if not exists contest (
    contest_id   integer not null primary key autoincrement,
    contest_name text    not null unique
    -- Fecha, organizador, 
);

create table if not exists problem (
    problem_id   integer not null primary key autoincrement,
    contest_id   integer,
    creator_id	 integer not null,
    problem_name text    not null,
    description  blob    not null,
    constraint fk_contest
	foreign key (contest_id)
	references contest(contest_id),
    constraint fk_creator
	foreign key (creator_id)
	references user(user_id)
);

create table if not exists test_case (
    problem_id     integer not null,
    num_test_case   integer not null,
    expected_out    blob    not null,
    given_input	    blob    not null,
    constraint fk_problem
	foreign key (problem_id)
	references problem(problem_id)
);

create table if not exists submission (
    submission_id   	  integer not null primary key autoincrement,
    user_id    	  	  integer not null,
    submission_status_id  integer not null,
    num_test_case 	  integer not null,
    problem_id   	  integer not null,
    constraint fk_user
	foreign key (user_id)
	references user(user_id),
    constraint fk_submission_status
	foreign key (submission_status_id)
	references submission_status(submission_status_id),
    constraint fk_num_test_case
	foreign key (num_test_case)
	references test_case(num_test_case),
    constraint fk_problem
        foreign key (problem_id)
	references problem(problem_id)
);

