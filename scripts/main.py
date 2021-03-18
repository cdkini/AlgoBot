#!/usr/bin/env python3

import requests

import firebase_admin
from firebase_admin import credentials
from firebase_admin import firestore

import tags, psets


API_BASE_URL = "https://leetcode.com/api/problems/algorithms/"
PROBLEM_BASE_URL = "https://leetcode.com/problems/"
PSETS = [
    "Top 100 Liked",
    "Top Interview",
    "Blind 75",
    "LeetCode Patterns",
]


class Question:
    def __init__(self, id, name, url, difficulty):
        self.id = id
        self.name = name
        self.url = url
        self.difficulty = difficulty
        self.tags = []
        self.psets = []

    def __repr__(self):
        return f"{self.id} - {self.name} ({self.difficulty})"

    def jsonify(self):

        def camel_case(item):
            if isinstance(item, list):
                return [camel_case(i) for i in item]
            elif isinstance(item, str):
                components = item.split(' ')
                return components[0].lower() + ''.join(x.title() for x in components[1:])
            else:
                return item

        for key, val in self.__dict__.items():
            if key == "name":
                continue
            self.__dict__[key] = camel_case(val)

        return self.__dict__


def get_all_questions():
    r = requests.get(API_BASE_URL)
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
    tags.add_tags(questions)
    psets.add_psets(questions)


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
