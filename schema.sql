-- These are the tables that make up our database

create table if not exists user (
    user_id   	 integer not null primary key autoincrement,
    username	 text    not null unique,
    passhash	 text    not null,
    email	 text    not null unique
);

create table if not exists status (
    status_id	integer	   not null primary key autoincrement,
    status	text	   not null unique
);

create table if not exists contest (
    contest_id	    	 integer not null primary key autoincrement,
    contest_name 	 text    not null unique,
    start_date 	 	 text	 not null,
    end_date   	 	 text	 not null,
    organizer_id 	 integer not null,
    constraint fk_organizer
        foreign key (organizer_id)
	references user(user_id)
);

-- Not useful right now
create table if not exists contest_has_problem (
    contest_has_problem_id integer not null primary key autoincrement,
    contest_id		   integer not null,
    problem_id		   integer not null,
    score		   integer not null,
    constraint fk_contest
        foreign key (contest_id)
	references contest(contest_id),
    constraint fk_problem
        foreign key (problem_id)
	references problem(problem_id)
);

create table if not exists problem (
    problem_id   integer not null primary key autoincrement,
    creator_id	 integer not null,
    problem_name text    not null,
    base_score	 integer not null,
    description  blob    not null,
    constraint fk_creator
	foreign key (creator_id)
	references user(user_id)
);

create table if not exists test_case (
    test_case_id   integer not null primary key autoincrement,
    problem_id     integer not null,
    num_test_case   integer not null,
    time_limit      integer not null,
    expected_out    blob    not null,
    given_input	    blob    not null,
    constraint fk_problem
	foreign key (problem_id)
	references problem(problem_id)
);

create table if not exists submission (
    submission_id   	  integer not null primary key autoincrement,
    user_id    	  	  integer not null,
    status_id  		  integer not null,
    score		  integer not null,
    date 	 	  text	 not null,
    problem_id   	  integer not null,
    constraint fk_user
	foreign key (user_id)
	references user(user_id),
    constraint fk_status
	foreign key (status_id)
	references status(status_id),
    constraint fk_problem
        foreign key (problem_id)
	references problem(problem_id)
);

create table if not exists test_case_status (
    test_case_status_id	   integer not null primary key autoincrement,
    submission_id	   integer not null,
    test_case_id	   integer not null,
    status_id		   integer not null,
    constraint fk_submission
        foreign key (submission_id)
	references submission(submission_id),
    constraint fk_test_case
        foreign key (test_case_id)
	references test_case(test_case_id),
    constraint fk_status
        foreign key (status_id)
	references status(status_id)
);

