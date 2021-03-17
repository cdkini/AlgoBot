import requests
from bs4 import BeautifulSoup


def add_psets(questions):
    top_100_liked(questions)
    top_interview(questions)
    print(f"Successfully assigned all questions under supported psets")


def top_100_liked(questions):
    path = "psets/top-100-liked.txt"
    with open(path) as f:
        content = [c.strip() for c in f.readlines()]
        for row in content:
            if row.isnumeric() and int(row) in questions:
                questions[int(row)].psets.append("Top 100 Liked")


def top_interview(questions):
    path = "psets/top-interview.txt"
    with open(path) as f:
        content = [c.strip() for c in f.readlines()]
        for row in content:
            if row.isnumeric() and int(row) in questions:
                questions[int(row)].psets.append("Top Interview")

