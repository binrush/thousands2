thousands.su - is a social network for summits climbers in the South Urals region. It contains a catalog of summits in South Urals higher then 1000 meters and an ability for users to create an account and register their climbs to these summits. The ultimate goal for users is to visit all summits higher then 1000 meters

There are approximately 300 summits in the catalog
A user can only register one climb for one summit (what matters is the number of visited summits, not the number of climbs)

# User stories
1. As a user, I want to be able to view summits catalog in a convenient way, including searching, sorting and viewing summits
2. As a user, I want to be able to view a detailed information about a specific summit and users who climbed it
3. As a user, I want to be able to register in the website using social networks
4. As a user, I want to be able to register my climb to a chosen summit and provide the date of climb and a comment about my climb
5. As a user, I want to be able to edit and delete my existing climbs
6. As a user, I want to be able to view a rating of climbers sorted by number of climbs
7. As a user, I want to be able to view other user's climbs
8. As a user, I want to use the website on all types of devices, including smartphones, tablets and desktop computers

# Website structure

Website should have a menu on top with the links to main pages: Summits catalog, Top
Website should have a link to login form in the top menu
For authenticated user, there should be a link to user profile page and logout button in the main menu

The website consites of the following pages

## Login form
Login form should have one button pointing to the Oauth login endpoint
It should be a modal window that can be opened on any page

## User profile page
URL: `/users/me`

The user profile page should contain user avatar, user name and the list of summits that the currently authenticated user has climbed
Every summit in the list should contain name, ridge, climb date and climb comment
The summit name in the list should be a link to a summit page
If the authenticated user opens his own user page, every element in the list should have controls to delete or edit a climb. The Edit control should direct user to the climb form (see below)

## Summits catalog
URL: `/summits`

A searchable and sorable list of summits, which includes summit name, summit height, ridge name which summit belongs,
a number of users that registered a climb to this summit, and a flag showing if this summit was visited by current  user
(in case user is logged in). 
Every summit in the list should be a link to this summit page (see below)
This page should be accessable for both authentiated and anonymous users

## Summit page
URL: `/{ridge_id}/{summit_id}`

The summit page should contain extended information about a specifc summit and a list of users who visited this summit on one page.

### Extended summit information
Should contain summit photo, name, alternative names, height, coordinates and description
If currently authenticated user did not register a climb yet, there should be a button or link pointing to the climb registration form (see below)
If current user us anonymous, the button shold open the login form

### List of users climbed the summit
Every user in the list should contain avatar, name, a date of climb and climb comment (if exists)
Every user in the list should be a link pointing to that user page (see below)
The list should be sorted by climb date (recent climbs first)
If the current user has a registered climb to this summit, he should appear as the first element of the list
The user list should have paging to prevent loading too much elements on the page

## User page
URL: `/user/{user_id}`

Same as user profile page, but without controls

## Climb form
URL: `/{ridge_id}/{summit_id}/climb`

The climb form is used to create or edit a climb. It should contain the following fields:
  * Climb date
  * Climb comment
Once form is submitted, the user should be redirected to his profile page

## Top page
Contains a list of top climbers. For every climber, his name, avatar, number of climbs and plate in the rating is displayed
Every list element should contain a link to the corresponding user page
The list should support paging to avoid overloading the page
