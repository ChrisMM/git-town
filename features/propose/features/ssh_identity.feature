@skipWindows
Feature: use a SSH identity

  Scenario Outline: ssh identity
    And tool "open" is installed
    And the origin is "git@my-ssh-identity:git-town/git-town.git"
    And Git Town setting "code-hosting-origin-hostname" is "<ORIGIN_HOSTNAME>"
    And the current branch is a feature branch "feature"
    When I run "git-town propose"
    Then "open" launches a new proposal with this url in my browser:
      """
      <PROPOSAL_URL>
      """

    Examples:
      | ORIGIN_HOSTNAME | PROPOSAL_URL                                                                                                                             |
      | bitbucket.org   | https://bitbucket.org/git-town/git-town/pull-requests/new?source=feature&dest=git-town%2Fgit-town%3Amain                                 |
      | github.com      | https://github.com/git-town/git-town/compare/feature?expand=1                                                                            |
      | gitlab.com      | https://gitlab.com/git-town/git-town/-/merge_requests/new?merge_request%5Bsource_branch%5D=feature&merge_request%5Btarget_branch%5D=main |
