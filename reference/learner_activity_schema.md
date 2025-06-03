LearnerActivity {
    sftpId	string
    userId	string
    firstName	string
    lastName	string
    audience	string
    contentUuid	string
    contentId	string
    contentTitle	string
    contentType	string Enum: [ AudioBook, Audio Summary, Book, Book Summary, Channel, Course, Journey, Linked Content, Video, Scheduled Content, Assessment ]
    status	string Enum: [ Achieved, Active, Completed, Listened, Read, Started, Watched ]
    completedDate	string
    duration	integer (units in seconds)
    firstAccess	string (ISO-8601 date format)
    lastAccess	string (ISO-8601 date format)
    totalAccesses	integer
    firstScore	integer
    highScore	integer
    lastScore	integer
    assessmentAttempts	integer
    % ofVideoOrBook	string (percentage of video or book completed)
    emailAddress	string
    languageCode	string
    customAttributes {
        description: Organization defined key/value pairs
        < * >:	string
                maxLength: 255
                nullable: true
        (example: OrderedMap { "Department": "IT - Northeast", "Employee Code": "12345" })
    } 
}