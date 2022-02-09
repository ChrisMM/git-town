Feature: offline mode

  Background:
    Given offline mode is enabled
    And the feature branches "feature" and "other"
    And the commits
      | BRANCH  | LOCATION      | MESSAGE        |
      | feature | local, origin | feature commit |
      | other   | local, origin | other commit   |
    And the current branch is "feature"
    And my workspace has an uncommitted file
    When I run "git-town kill"

  Scenario: result
    Then it runs the commands
      | BRANCH  | COMMAND                        |
      | feature | git add -A                     |
      |         | git commit -m "WIP on feature" |
      |         | git checkout main              |
      | main    | git branch -D feature          |
    And the current branch is now "main"
    And my repo doesn't have any uncommitted files
    And the branches are now
      | REPOSITORY | BRANCHES             |
      | local      | main, other          |
      | origin     | main, feature, other |
    And now these commits exist
      | BRANCH  | LOCATION      | MESSAGE        |
      | feature | origin        | feature commit |
      | other   | local, origin | other commit   |
    And Git Town is now aware of this branch hierarchy
      | BRANCH | PARENT |
      | other  | main   |

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH  | COMMAND                                       |
      | main    | git branch feature {{ sha 'WIP on feature' }} |
      |         | git checkout feature                          |
      | feature | git reset {{ sha 'feature commit' }}          |
    And the current branch is now "feature"
    And my workspace has the uncommitted file again
    And now the initial commits exist
    And the initial branches and hierarchy exist
