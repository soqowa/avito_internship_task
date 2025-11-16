import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const teamRes = http.post(`${BASE_URL}/teams`, JSON.stringify({ name: 'team-k6' }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(teamRes, { 'team created': (r) => r.status === 201 });
  const team = teamRes.json();

  const users = [];
  for (let i = 0; i < 3; i++) {
    const uRes = http.post(
      `${BASE_URL}/teams/${team.id}/users`,
      JSON.stringify({ name: `user-${i}` }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    check(uRes, { 'user created': (r) => r.status === 201 });
    users.push(uRes.json());
  }

  return { team, users };
}

export default function (data) {
  const author = data.users[0];

  const prRes = http.post(
    `${BASE_URL}/prs`,
    JSON.stringify({ title: 'k6-pr', authorId: author.id }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  check(prRes, { 'pr created': (r) => r.status === 201 });
  const pr = prRes.json();

  if (pr.reviewers && pr.reviewers.length > 0) {
    const oldReviewer = pr.reviewers[0].userId;
    const reassignRes = http.post(
      `${BASE_URL}/prs/${pr.id}/reassign`,
      JSON.stringify({ oldReviewerId: oldReviewer }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    check(reassignRes, {
      'reassign ok or conflict': (r) => r.status === 200 || r.status === 409,
    });
  }

  const mergeRes = http.post(`${BASE_URL}/prs/${pr.id}/merge`, null);
  check(mergeRes, { 'merge ok': (r) => r.status === 200 });

  sleep(1);
}

