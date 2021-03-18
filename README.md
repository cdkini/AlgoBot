# AlgoBot :snake::robot:
<a href='http://www.recurse.com' title='Made with love at the Recurse Center'><img src='https://cloud.githubusercontent.com/assets/2883345/11325206/336ea5f4-9150-11e5-9e90-d86ad31993d8.png' height='20px'/></a>

A Zulip bot made for [The Recurse Center](http://www.recurse.com) to help individuals get better at data structures and algorithms!

AlgoBot has three primary features:

1. [Solo Sessions](#solo-sessions) - Setting up a study schedule and providing daily questions through Zulip private message.
2. [Mock Interviews](#mock-interviews) - Pairing similarly experienced individuals for mock interviews.
3. [Daily Questions](#daily-questions) - Posting daily questions to Zulip threads for the community to work on together.

Please note that AlgoBot is a direct fork of [thwidge/pairing-bot](https://github.com/thwidge/pairing-bot).
This project would not have been possible without the hard work put in by Maren Beam and the rest of the contributors over there. AlgoBot says thank you :pray:

<hr>

## Table of Contents
1. [Usage](#usage)
    1. [Commands](#commands)
    2. [Solo Sessions](#solo-sessions)
    3. [Mock Interviews](#mock-interviews)
    4. [Daily Questions](#daily-questions)
2. [Design](#design)
    1. [Stack](#stack)
    3. [Matching Algorithm](#matching-algo)
    2. [Deployment](#deployment)
3. [Contributing](#contributing)
    1. [Pull Requests](#pull-requests)
4. [Mock Interviews](#mock-interviews)
5. [FAQ](#faq)
    1. [Why do I need to add myself back to the mock interview queue after each session?](#faq-mock-interview-queue)
    2. [I'm a beginner to these types of problems. Can I still pair?](#faq-beginner)
    3. [Can I just practice my interviewer skills without being interviewed?](#faq-interviewer)
    4. [Why should I use AlgoBot over XYZ?](#faq-alternatives)
    5. [Do I need to code in Python?](#faq-python)
6. [Additional Resources](#resources)
    1. [Mock Interviews](#resources-mock-interviews)
    2. [Additional Questions](#resources-additional-questions)
    3. [DS&A Courses / MOOCS](#resources-courses)

<hr>

<a name="usage"></a>
## 1. Usage 

<a name="commands"></a>
### 1.i. Commands 

Upon typing `help` in a private chat with AlgoBot, you'll receive the a response informing you of the following supported commands:

- `subscribe` to get started! You'll start getting daily data structures and algorithms questions. Oh how fun!
- `schedule` to add yourself to the queue for a mock interview!
  - You'll remain in the queue until you get a match! Upon interviewing, you'll need to `schedule` once again.
  - In the case you no longer can mock interview, please `cancel`.
- `skip` to skip tomorrow's daily question.
  - `unskip` if you change your mind.
- `config` to review and modify your current settings
- `unsubscribe` to part ways with AlgoBot. Note that your settings and session history will be deleted!
 
Note that these commands only work in a 1-on-1 chat with AlgoBot.

<a name="solo-sessions"></a>
### 1.ii. Solo Sessions

The bread and butter of AlgoBot, solo sessions are the questions you receive each day as part of your structured study plan.
Questions will be selected from either a problem set or at random based on your configuration. Feel free to treat these as seriously as you'd like;
it's entirely up to you whether you treat them as serious interview prep, a fun exercise to work on with friends, or something in between.

Upon using the `subscribe` cmd, your account will be assigned the following default configurations:
- `Days`: Mon/Tue/Wed/Thu/Fri
- `Difficulty`: Easy / Medium (randomly selected between the two)
- `Topics`: All / Random
- `Problem Set`: Top Interview Questions (LeetCode)

These defaults can be viewed and altered at any time using the `config` option. 
Note that questions are sent out at 9:00AM EST on the scheduled day so any changes or `skip` cmds will need to be made before then.


<a name="mock-interviews"></a>
### 1.iii. Mock Interviews 

When you decide that you want to take a stab at actual interviews, use mock interviews to get in that practice with other Recursers.
Questions will be selected from either a problem set or by you manually; if you pick your own question, make sure to let your partner know!
These sessions are meant to simulate real world interviews; treat them like phone screens and you'll find success!

Upon using the `schedule` cmd, you will be placed in the queue with other Recursers interested in mock interviewing. You will be assigned the following defaults:
- `Days`: You're in the queue until you are matched; upon matching, discuss possible availability your partner.
- `Experience`: Medium (you will prepare questions of medium difficulty and below as an interviewer). 
- `Difficulty`: Easy / Medium (randomly selected between the two)
- `Topics`: All / Random
- `Problem Set`: Top Interview Questions (LeetCode)
- `Environment`: LeetCode

These defaults can be viewed and altered at any time using the `config` option. 
Note that matches are made and sent out at 11:00PM EST each day so any changes or `cancel` cmds will need to be made before then.

It is important to note that matches are made based on similarity of profiles to ensure equitable, rewarding interviews.
Please take the time to review your configuration to ensure it matches your preferences and experience level.

If you do not get a match the first day, do not worry! You'll stay in the queue until you do. Upon matching, you'll be removed from the queue. 
If you want back-to-back interviews, you'll need to manually `schedule` all over again. Read the [FAQ](#faq-mock-interview-queue) for more details.

<a name="daily-questions"></a>
### 1.iv. Daily Questions

For those that want to treat these questions as a collaborative effort, daily questions are posted to a Zulip thread.
The difficulty of these questions increases throughout the week (akin to something like the NYT crossword).

| Day    | Difficulty  |
| ------ | ----------- |
| Mon    | Easy        |
| Tue    | Easy        |
| Wed    | Medium      |
| Thu    | Medium      |
| Fri    | Medium      |
| Sat    | Hard        |
| Sun    | Hard        |


Link to thread: [Daily Question](link)

Note that you do not require any special configuration or messaging of AlgoBot for these problems. 
Simply go to the link, take a stab at the question, and post your solution in the thread to discuss with others.
Remember to use spoiler tags to prevent ruining the solution for others!

<hr>

<a name="design"></a>
## 2. Design 

<a name="stack"></a>
### 2.i. Stack

<a name="matching-algo"></a>
### 2.ii. Matching Algorithm

<a name="deployment"></a>
### 2.iii. Deployment 

### About Pairing Bot's setup and deployment
 * Serverless. RC's instance is currently deployed on [App Engine](https://cloud.google.com/appengine/docs/standard/)
 * [Firestore database](https://cloud.google.com/firestore/docs/)
 * Deployed on pushes to the `main` branch with [Cloud Build](https://cloud.google.com/cloud-build/docs/)
 * The database must be prepopulated with two pieces of data:  an authentication token (which the bot uses to validate incoming webhook requests), and an api key (which the bot uses to send private messages to Zulip users)
 * Zulip has bot types. Pairing Bot is of type `outgoing webhook`
 * Pair programming matches are made, and the people who've been matched are notified, any time an HTTP GET request is issued to `/cron`

<hr>

<a name="contributing"></a>
## 3. Contributing 

<a name="contributions"></a>
### 3.i. What you can do to help

This is meant to be a project built by and for the Recurse Center community so feel free to contribute!

I'll be posting issues and/or feature ideas to the issues board as they arise so take a look there. 
If you've found a bug or think of a feature that would improve the AlgoBot experience, feel free to add your own issue.
Before actually contributing, comment on the issue you're tackling and I'll assign it to you.

In the case you can't find an issue, bug, or feature to work on, here's a list of ideas (listed by language):

| Language    | Contribution                                                            |
| ------      | -----------                                                             |
| Go          | Tests / documentation                                                   |
| Python      | Tests / documentation / scraping new psets / improving scraping scripts |
| HTML/CSS/JS | Improve form validation / add general styling                           |
| Misc        | Documentation                                                           |


**Your contributions are welcome and encouraged, no matter your prior experience!**

<a name="pull-requests"></a>
### 3.ii. Pull Requests

The workflow below has proven to be useful with other projects but please let me know if I can clarify anything!

```
1. Create an issue
2. Fork the repo
3. Create a branch*
4. Make your changes
5. Write unit tests as applicable
6. Format the codebase using 'go fmt'*
7. Ensure that your changes passes all tests using 'go test'
8. Squash your changes to as few commits as possible*
9. Make a pull request*
```
<i>\*Please use the issue number and name when possible to improve clarity and project maintainability. 
<br>Additionally, please adhere to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) standards.</i>

<hr>

<a name="mock-interviews"></a>
## 4. Mock Interviews 

<a name="as-interviewer"></a>
### 4.i. What should I expect or prepare for as the interviewer?

Upon getting matched, your responsibility as the interviewer is to prepare 

If your interviewee seems to be stuck or going in the wrong direction, 

<a name="as-interviewee"></a>
### 4.ii. What should I expect or prepare for as the interviewee?

If your interviewee seems to be stuck or going in the wrong direction, 

<hr>

<a name="faq"></a>
## 5. FAQ

<a name="faq-mock-interview-queue"></a>
### 5.i. Why do I need to add myself back to the mock interview queue after each session?
 
I decided that it's best that people opt in each time they want to interview to prevent the situation where people forget and possibly ruin the experience for their partner.
I think the decision makes sense since interviewing is less common and frequent than something like pairing but I'd be happy to listen to any alternatives y'all might have!

<a name="faq-beginner"></a>
### 5.ii. I'm a beginner to these types of problems. Can I still pair?

Most definitely! A big part of data structures and algorithms problems is repetition and pattern recognition; how are you going to recognize the patterns without some exposure?
The only thing I ask of you is to change your configuration to reflect your current experience level so you can be paired with someone of a similar experience level (using `config`). 
I'd also highly recommend taking a course or MOOC about the topic if you have the time (see 5.iv. for recommendations)! In the case you just aren't understanding the question you're supposed to prepare as the interviewer, talk to your partner and see if you can come up with an alternative. 

<a name="faq-interviewer"></a>
### 5.iii. Can I just practice my interviewer skills without being interviewed?

Bless your heart! Just set your configuration to match the experience level you wish to interview and let your partner know that they don't need to prepare anything. You sound like a good interviewer :)

<a name="faq-alternatives"></a>
### 5.iv. Why should I use AlgoBot over XYZ?

You should not! AlgoBot is a <b>supplement</b> to the rest of your prep and a great way to get to work with your fellow Recursers in a bit of a different setting.
Check out the additional resources I've linked below.

<a name="faq-python"></a>
### 5.v. Do I need to code in Python :snake: :snake: :snake: ?

I'll be very sad if you don't but disregard my feelings and use whatever language best fits your preferences.
In the case your interviewer doesn't know the language you're using, try to explain language-specific features as you use them.
I thought of adding a preferred language as a configuration feature but ultimately decided against it as it would further splinter participants and make matching harder.

<hr>

### 6. Additional Resources

#### 6.i. Mock Interviews
  - [Pramp](https://www.pramp.com)
    - The primary inspiration for this tool, Pramp is completely free and great for getting paired with similarly experienced devs prepping for interviews. I've heard the question pool is a bit limited but it's been solid in my experience.
  - [Recurse Career Services](https://www.recurse.com/jobs/advice#interviews)
    - Your best bet for a realistic interview. Talk to the team and I'm sure they'll get you a great match!

#### 6.ii. Additional Questions (ranked by easiest to hardest)
  - [Codewars](https://www.codewars.com/)
    - Not entirely focused on DS&A but I've found it to be a great intro to small, challenging problems.
  - [Firecode.io](https://www.firecode.io)
    - A very solid resource that uses repetition to ensure you have a grasp on certain concepts before moving forward. The one caveat is that it supports Java/C/C++ and not Python. Hsssss :snake:! 
  - [CodeSignal](https://codesignal.com)
    - Somewhat similar to Codewars but it has a great section on common DS&A problems.
  - [interviews.school](https://interviews.school)
    - Just a curated list of LeetCode problems but it does a good job of listing out what you need to know.
  - [Advent of Code](https://www.adventofcode.com)
    - I love me some AoC. While these problem don't explicitly require efficient solutions, I often code up the brute force solution and then improve upon them. The subreddit has some real coding wizards.
  - [HackerRank](https://www.hackerrank.com)
    - Pretty much on par with LeetCode but does require you to read in input! Some online assessments and phone screens are done here so might be worth checking out!
  - [LeetCode](https://www.leetcode.com)
    - Hello darkness my old friend. AlgoBot questions are pulled from here so you know it's good.
     
#### 6.iii. DS&A Courses / MOOCS
  - [Cracking the Coding Interview](https://cin.ufpe.br/~fbma/Crack/Cracking%20the%20Coding%20Interview%20189%20Programming%20Questions%20and%20Solutions.pdf)
    - Not a course but the standard recommendation (and for good reason). I think its significantly easier than LeetCode but worth checking out. Yer a wizard Gayle Laakmann McDowell!
  - [Princeton Algorithms](https://online.princeton.edu/node/201)
    - I've only done a few psets from this course and have dabbled in parts of the textbook but it's definitely some good stuff. Coursera constantly runs free courses for both parts!
  - [Berkeley CS61B](https://sp18.datastructur.es/)
    - Josh Hug is one of the greatest professors I've had the pleasure of learning from (also shout out to David Malan). This course has a fair amount of overlap with Sedgewick's course since Prof. Hug taught at Princeton for a bit.
      Some of the projects, like generating random maps for a Dwarf Fortress-esque game or writing your own Git, were incredibly fun and challenging. The Spring 2018 version released its autograder so you can treat it like a real course!
  - [MIT 6.006](https://ocw.mit.edu/courses/electrical-engineering-and-computer-science/6-006-introduction-to-algorithms-fall-2011/)
    - I've only watched lectures from this course but Eric Demain and Srini Devadas are incredible at what they do. 
    Some of the material might be a bit niche for interviews and the lectures can get a bit "mathy" but if you're looking to strengthen your understanding of DS&A, look no further! 
