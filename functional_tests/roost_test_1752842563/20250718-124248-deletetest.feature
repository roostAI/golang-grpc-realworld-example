Feature: User Login System

Background:
  Given the API base URL is '${BASE_URL}/api/v1'
  And the content type header is set to 'application/json'
  And the request timeout is set to 30 seconds

# TC-001: Basic Login Functionality
Scenario: Successful user login with valid credentials
  Given a valid user account exists in the system
  When I send a POST request to "/login" with body:
    """
    {
      "username": "testuser@example.com",
      "password": "Password123!"
    }
    """
  Then the response status code should be 200
  And the response should contain an authentication token
  And the response should contain "userId"
  And the response time should be less than 1000 milliseconds
  And the response headers should contain "Set-Cookie"

# TC-002: Invalid Login Attempt
Scenario Outline: Failed login attempts with invalid credentials
  When I send a POST request to "/login" with body:
    """
    {
      "username": "<username>",
      "password": "<password>"
    }
    """
  Then the response status code should be 401
  And the response body should contain "error": "<error_message>"
  And the response should not contain an authentication token

  Examples:
    | username              | password      | error_message                    |
    | wrong@example.com     | WrongPass123  | Invalid credentials             |
    | testuser@example.com  | wrongpass     | Invalid credentials             |
    |                      | Password123!   | Username is required            |
    | testuser@example.com  |               | Password is required            |

# TC-003: Login Page Load Time Performance
Scenario: Verify login endpoint response time
  Given the system is operational
  When I send a GET request to "/login"
  Then the response status code should be 200
  And the response time should be less than 3000 milliseconds
  And the response headers should contain "Cache-Control"

# TC-004: Login Security Measures
Scenario: Verify secure communication for login
  Given the system is operational
  When I send a POST request to "/login" with body:
    """
    {
      "username": "testuser@example.com",
      "password": "Password123!"
    }
    """
  Then the response status code should be 200
  And the response headers should contain "Strict-Transport-Security"
  And the response headers should contain "X-XSS-Protection"
  And the response headers should contain "X-Content-Type-Options"
  And the response cookies should have "Secure" flag
  And the response cookies should have "HttpOnly" flag
  And the response cookies should have "SameSite" attribute

Scenario: Verify rate limiting for login attempts
  Given I have made 5 failed login attempts
  When I send a POST request to "/login" with body:
    """
    {
      "username": "testuser@example.com",
      "password": "WrongPassword"
    }
    """
  Then the response status code should be 429
  And the response body should contain "error": "Too many login attempts"
  And the response headers should contain "Retry-After"

Scenario: Verify password field encryption
  Given I capture the network traffic
  When I send a POST request to "/login" with body:
    """
    {
      "username": "testuser@example.com",
      "password": "Password123!"
    }
    """
  Then the password should be transmitted in encrypted format
  And the response status code should be 200
