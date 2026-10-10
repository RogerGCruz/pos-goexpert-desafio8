Feature: Auction automatic closing
  Ensure auctions are closed automatically after configured duration

  Background:
    Given the app is running

  Scenario: Auction closes automatically after duration
    When I create an auction with product name "BDD Product" and category "BDD" and description "A product for bdd tests" and condition 1
    And I search for active auctions with product name "BDD Product"
    Then I should find at least one auction
    When I wait for the auction duration
    And I fetch the auction by id
    Then the auction status should be 1
