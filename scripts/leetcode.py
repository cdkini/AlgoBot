#!/usr/bin/env python3

import os
import requests

import firebase_admin
from firebase_admin import credentials
from firebase_admin import firestore


RAW_DATA = "https://leetcode.com/api/problems/algorithms/"
PROBLEM_BASE_URL = "https://leetcode.com/problems/"
TAGS = [
    "Array",
    "Backtracking",
    "Binary Search",
    "Bit Manipulation",
    "Breadth-first Search",
    "Depth-first Search",
    "Design",
    "Divide and Conquer",
    "Dynamic Programming",
    "Graph",
    "Greedy",
    "Hash Table",
    "Heap",
    "Linked List",
    "Math",
    "Recursion",
    "Sliding Window",
    "Sort",
    "Stack",
    "String",
    "Tree",
    "Trie",
    "Two Pointers",
    "Union Find"
]


class Question:
    def __init__(self, id, name, url, difficulty):
        self.id = id
        self.name = name
        self.url = url
        self.difficulty = difficulty
        self.tags = []

    def __repr__(self):
        return f"{self.id} - {self.name} ({self.difficulty})"

    def jsonify(self):
        return {
            'id': self.id,
            'name': self.name,
            'url': self.url,
            'difficulty': self.difficulty,
            'tags': self.tags 
        }


def get_all_questions():
    r = requests.get(RAW_DATA)
    if r.status_code == 200:
        print("Successfully hit LeetCode API endpoint")
    return r.json()


def clean_raw_data(data):
    questions = data.get("stat_status_pairs")
    difficulties = ["Easy", "Medium", "Hard"]

    clean = {}
    for question in questions:
        if question.get("paid_only"):
            continue
        stat = question.get("stat")

        id = stat.get("frontend_question_id")
        name = stat.get("question__title")
        url = PROBLEM_BASE_URL + stat.get('question__title_slug')
        difficulty = difficulties[question.get("difficulty").get("level") - 1]

        question = Question(id, name, url, difficulty)
        clean[id] = question

    print(f"Successfully serialized {len(clean)} questions as Python objects")
    return clean


def classify_questions(questions):
    for tag in TAGS:
        path = _get_tag_path(tag)
        with open(path) as f:
            content = [c.strip() for c in f.readlines()]
            for row in content:
                if row.isnumeric() and int(row) in questions:
                    questions[int(row)].tags.append(tag)

    print(f"Successfully classified all questions under {len(TAGS)} tags")


def _get_tag_path(tag):
    arr = tag.split(" ")
    joined = "-".join(word.lower() for word in arr)
    return f"tags/{joined}.txt"


def populate_firebase(questions):
    client = _init_gcloud()
    for id, question in questions.items():
        doc_ref = client.collection("questions").document(str(id))
        doc_ref.set(question.jsonify())
    print(f"Successfully updated {len(questions)} documents in Firebase")


def _init_gcloud():
    cred = credentials.Certificate('secrets.json')
    firebase_admin.initialize_app(cred)
    print("Successfully authorized and connected to Firebase instance")
    return firestore.client()
     

def main():
    try:
        raw = get_all_questions()
        clean = clean_raw_data(raw)
        classify_questions(clean)
        populate_firebase(clean)
        print("SUCCESS: Database successfully populated with LeetCode data")
    except:
        print("FAILURE: An exception occurred somewhere")


if __name__ == "__main__":
    main()
